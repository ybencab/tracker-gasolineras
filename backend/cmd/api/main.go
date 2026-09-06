package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"tracker-gasolineras-backend/internal/cache"
	"tracker-gasolineras-backend/internal/geoportal"
)

func main() {
	loadDotEnv(".env")

	baseURL := os.Getenv("GEOPORTAL_BASE_URL")
	if baseURL == "" {
		baseURL = geoportal.DefaultBaseURL
	}

	refreshInterval := 24 * time.Hour
	if v := os.Getenv("REFRESH_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Fatalf("REFRESH_INTERVAL inválido (%q): %v", v, err)
		}
		refreshInterval = d
	}

	snapshotPath := os.Getenv("SNAPSHOT_PATH")
	if snapshotPath == "" {
		snapshotPath = "./data/snapshot.json"
	}

	client := geoportal.NewClient(baseURL)
	store := cache.NewStore(client, snapshotPath)

	if err := store.LoadFromDisk(); err != nil {
		log.Printf("sin snapshot previo en disco (%v); se servirá en cuanto termine el primer refresco", err)
	}

	go store.Start(context.Background(), refreshInterval)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/provincias", func(w http.ResponseWriter, r *http.Request) {
		provincias, ok := store.Provincias()
		if !ok {
			writeSinDatos(w)
			return
		}
		writeJSON(w, provincias)
	})

	mux.HandleFunc("GET /api/municipios", func(w http.ResponseWriter, r *http.Request) {
		idProvincia := r.URL.Query().Get("provincia")
		if idProvincia == "" {
			http.Error(w, "falta el parámetro 'provincia'", http.StatusBadRequest)
			return
		}
		municipios, ok := store.MunicipiosPorProvincia(idProvincia)
		if !ok {
			writeSinDatos(w)
			return
		}
		writeJSON(w, municipios)
	})

	mux.HandleFunc("GET /api/estaciones", func(w http.ResponseWriter, r *http.Request) {
		idMunicipio := r.URL.Query().Get("municipio")
		if idMunicipio == "" {
			http.Error(w, "falta el parámetro 'municipio'", http.StatusBadRequest)
			return
		}
		fecha, estaciones, ok := store.EstacionesPorMunicipio(idMunicipio)
		if !ok {
			writeSinDatos(w)
			return
		}
		writeJSON(w, geoportal.EstacionesResponse{Fecha: fecha, Estaciones: estaciones})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// withCORS permite las llamadas desde el frontend, servido en otro origen
// durante el desarrollo (Astro en :4321, API en :8080).
func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// writeSinDatos se usa mientras el primer refresco en background todavía
// no ha terminado (y no había snapshot previo en disco).
func writeSinDatos(w http.ResponseWriter) {
	http.Error(w, "aún no hay datos cargados, inténtalo de nuevo en unos segundos", http.StatusServiceUnavailable)
}

// loadDotEnv carga variables desde un fichero .env (si existe) sin
// sobrescribir las que ya estén definidas en el entorno real.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		clave, valor, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		clave = strings.TrimSpace(clave)
		if _, definida := os.LookupEnv(clave); !definida {
			os.Setenv(clave, strings.TrimSpace(valor))
		}
	}
}
