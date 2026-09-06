// Package geoportal es un cliente para la API pública de consulta de precios
// de carburantes del Geoportal Gasolineras (MITECO).
package geoportal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL es la URL base pública del Geoportal Gasolineras,
// usada si no se define la variable de entorno GEOPORTAL_BASE_URL.
const DefaultBaseURL = "https://sedeaplicaciones.minetur.gob.es/ServiciosRESTCarburantes/PreciosCarburantes"

// El Geoportal solo acepta cifrado TLS DHE (sin ECDHE), que la librería
// estándar de Go no soporta. Por eso las peticiones se delegan en el
// binario `curl` del sistema (enlazado con OpenSSL) en vez de net/http.
type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

type Provincia struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	IDCCAA string `json:"id_ccaa"`
}

type Municipio struct {
	ID          string `json:"id"`
	Nombre      string `json:"nombre"`
	IDProvincia string `json:"id_provincia"`
}

type Estacion struct {
	ID           string             `json:"id"`
	IDMunicipio  string             `json:"id_municipio"`
	IDProvincia  string             `json:"id_provincia"`
	Rotulo       string             `json:"rotulo"`
	Direccion    string             `json:"direccion"`
	CodigoPostal string             `json:"codigo_postal"`
	Localidad    string             `json:"localidad"`
	Municipio    string             `json:"municipio"`
	Provincia    string             `json:"provincia"`
	Horario      string             `json:"horario"`
	Latitud      float64            `json:"latitud"`
	Longitud     float64            `json:"longitud"`
	Precios      map[string]float64 `json:"precios"`
}

type EstacionesResponse struct {
	Fecha      string     `json:"fecha"`
	Estaciones []Estacion `json:"estaciones"`
}

// rawProvincia refleja tal cual los campos que devuelve el Geoportal
// (incluido el typo "IDPovincia", real de la fuente).
type rawProvincia struct {
	IDPovincia string `json:"IDPovincia"`
	IDCCAA     string `json:"IDCCAA"`
	Provincia  string `json:"Provincia"`
}

type rawMunicipio struct {
	IDMunicipio string `json:"IDMunicipio"`
	IDProvincia string `json:"IDProvincia"`
	Municipio   string `json:"Municipio"`
}

type rawEstacionesResponse struct {
	Fecha           string              `json:"Fecha"`
	ListaEESSPrecio []map[string]string `json:"ListaEESSPrecio"`
}

// Provincias devuelve el listado completo de provincias.
func (c *Client) Provincias(ctx context.Context) ([]Provincia, error) {
	var raw []rawProvincia
	if err := c.getJSON(ctx, "/Listados/Provincias/", &raw); err != nil {
		return nil, err
	}
	provincias := make([]Provincia, 0, len(raw))
	for _, p := range raw {
		provincias = append(provincias, Provincia{
			ID:     p.IDPovincia,
			Nombre: p.Provincia,
			IDCCAA: p.IDCCAA,
		})
	}
	return provincias, nil
}

// Municipios devuelve el listado completo de municipios de España.
func (c *Client) Municipios(ctx context.Context) ([]Municipio, error) {
	var raw []rawMunicipio
	if err := c.getJSON(ctx, "/Listados/Municipios/", &raw); err != nil {
		return nil, err
	}
	municipios := make([]Municipio, 0, len(raw))
	for _, m := range raw {
		municipios = append(municipios, Municipio{
			ID:          m.IDMunicipio,
			Nombre:      m.Municipio,
			IDProvincia: m.IDProvincia,
		})
	}
	return municipios, nil
}

// Estaciones devuelve todas las estaciones de servicio de España con sus
// precios. Es una respuesta grande (~11.500 estaciones, ~12MB), pensada
// para pedirse una vez por refresco, no por petición de usuario.
func (c *Client) Estaciones(ctx context.Context) (EstacionesResponse, error) {
	var raw rawEstacionesResponse
	if err := c.getJSON(ctx, "/EstacionesTerrestres/", &raw); err != nil {
		return EstacionesResponse{}, err
	}

	estaciones := make([]Estacion, 0, len(raw.ListaEESSPrecio))
	for _, campos := range raw.ListaEESSPrecio {
		estaciones = append(estaciones, parseEstacion(campos))
	}

	return EstacionesResponse{
		Fecha:      raw.Fecha,
		Estaciones: estaciones,
	}, nil
}

func parseEstacion(campos map[string]string) Estacion {
	precios := make(map[string]float64)
	for clave, valor := range campos {
		nombre, esPrecio := strings.CutPrefix(clave, "Precio ")
		if !esPrecio {
			continue
		}
		if precio, ok := parseDecimal(valor); ok {
			precios[nombre] = precio
		}
	}

	latitud, _ := parseDecimal(campos["Latitud"])
	longitud, _ := parseDecimal(campos["Longitud (WGS84)"])

	return Estacion{
		ID:           campos["IDEESS"],
		IDMunicipio:  campos["IDMunicipio"],
		IDProvincia:  campos["IDProvincia"],
		Rotulo:       campos["Rótulo"],
		Direccion:    campos["Dirección"],
		CodigoPostal: campos["C.P."],
		Localidad:    campos["Localidad"],
		Municipio:    campos["Municipio"],
		Provincia:    campos["Provincia"],
		Horario:      campos["Horario"],
		Latitud:      latitud,
		Longitud:     longitud,
		Precios:      precios,
	}
}

// parseDecimal convierte el formato numérico español ("1,615") a float64.
// Devuelve ok=false si el valor está vacío (sin precio) o no es parseable.
func parseDecimal(valor string) (float64, bool) {
	if valor == "" {
		return 0, false
	}
	normalizado := strings.ReplaceAll(valor, ",", ".")
	numero, err := strconv.ParseFloat(normalizado, 64)
	if err != nil {
		return 0, false
	}
	return numero, true
}

func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "curl", "-sS", "--fail", c.baseURL+path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("geoportal: curl %s: %w (%s)", path, err, strings.TrimSpace(stderr.String()))
	}

	if err := json.Unmarshal(stdout.Bytes(), out); err != nil {
		return fmt.Errorf("geoportal: decodificando respuesta de %s: %w", path, err)
	}
	return nil
}
