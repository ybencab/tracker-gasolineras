// Package cache mantiene en memoria el último snapshot de datos del
// Geoportal Gasolineras, refrescado periódicamente en background, para que
// ninguna petición de usuario dependa de una llamada en vivo al Geoportal.
package cache

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tracker-gasolineras-backend/internal/geoportal"
)

type datos struct {
	Fecha      string                `json:"fecha"`
	Provincias []geoportal.Provincia `json:"provincias"`
	Municipios []geoportal.Municipio `json:"municipios"`
	Estaciones []geoportal.Estacion  `json:"estaciones"`
}

type Store struct {
	mu     sync.RWMutex
	datos  *datos
	client *geoportal.Client
	path   string
}

func NewStore(client *geoportal.Client, snapshotPath string) *Store {
	return &Store{client: client, path: snapshotPath}
}

// LoadFromDisk intenta cargar un snapshot previo, para poder servir algo
// inmediatamente al arrancar si aún no ha terminado el primer refresco.
func (s *Store) LoadFromDisk() error {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var d datos
	if err := json.Unmarshal(raw, &d); err != nil {
		return err
	}
	s.mu.Lock()
	s.datos = &d
	s.mu.Unlock()
	return nil
}

// Start lanza un refresco inmediato y, a partir de ahí, uno cada `interval`.
// Pensado para ejecutarse en su propia goroutine; bloquea hasta que ctx se
// cancela.
func (s *Store) Start(ctx context.Context, interval time.Duration) {
	if err := s.refresh(ctx); err != nil {
		log.Printf("cache: refresco inicial falló: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.refresh(ctx); err != nil {
				log.Printf("cache: refresco falló, se mantienen los datos anteriores: %v", err)
			}
		}
	}
}

func (s *Store) refresh(ctx context.Context) error {
	provincias, err := s.client.Provincias(ctx)
	if err != nil {
		return err
	}
	municipios, err := s.client.Municipios(ctx)
	if err != nil {
		return err
	}
	resp, err := s.client.Estaciones(ctx)
	if err != nil {
		return err
	}

	d := &datos{
		Fecha:      resp.Fecha,
		Provincias: provincias,
		Municipios: municipios,
		Estaciones: resp.Estaciones,
	}

	s.mu.Lock()
	s.datos = d
	s.mu.Unlock()

	if err := s.persist(d); err != nil {
		log.Printf("cache: no se pudo guardar el snapshot en disco: %v", err)
	}

	log.Printf("cache: refresco completado (%d provincias, %d municipios, %d estaciones, fecha origen %s)",
		len(provincias), len(municipios), len(resp.Estaciones), resp.Fecha)
	return nil
}

func (s *Store) persist(d *datos) error {
	raw, err := json.Marshal(d)
	if err != nil {
		return err
	}
	if dir := filepath.Dir(s.path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(s.path, raw, 0o644)
}

// Provincias devuelve el listado de provincias del snapshot actual.
// ok=false si todavía no hay ningún snapshot cargado.
func (s *Store) Provincias() (provincias []geoportal.Provincia, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.datos == nil {
		return nil, false
	}
	return s.datos.Provincias, true
}

// MunicipiosPorProvincia filtra en memoria los municipios de una provincia.
func (s *Store) MunicipiosPorProvincia(idProvincia string) (municipios []geoportal.Municipio, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.datos == nil {
		return nil, false
	}
	municipios = []geoportal.Municipio{}
	for _, m := range s.datos.Municipios {
		if m.IDProvincia == idProvincia {
			municipios = append(municipios, m)
		}
	}
	return municipios, true
}

// EstacionesPorMunicipio filtra en memoria las estaciones de un municipio,
// y devuelve también la fecha de los datos de origen.
func (s *Store) EstacionesPorMunicipio(idMunicipio string) (fecha string, estaciones []geoportal.Estacion, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.datos == nil {
		return "", nil, false
	}
	estaciones = []geoportal.Estacion{}
	for _, e := range s.datos.Estaciones {
		if e.IDMunicipio == idMunicipio {
			estaciones = append(estaciones, e)
		}
	}
	return s.datos.Fecha, estaciones, true
}
