package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"tracker-gasolineras-backend/internal/geoportal"
)

func main() {
	loadDotEnv(".env")

	baseURL := os.Getenv("GEOPORTAL_BASE_URL")
	if baseURL == "" {
		baseURL = geoportal.DefaultBaseURL
	}
	client := geoportal.NewClient(baseURL)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/provincias", func(w http.ResponseWriter, r *http.Request) {
		provincias, err := client.Provincias(r.Context())
		if err != nil {
			writeError(w, err)
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
		municipios, err := client.MunicipiosPorProvincia(r.Context(), idProvincia)
		if err != nil {
			writeError(w, err)
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
		estaciones, err := client.EstacionesPorMunicipio(r.Context(), idMunicipio)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, estaciones)
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

func writeError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, "error consultando el Geoportal Gasolineras", http.StatusBadGateway)
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
