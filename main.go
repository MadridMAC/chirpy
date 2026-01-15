package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	api_cfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	s_mux := http.NewServeMux()
	app_handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	s_mux.Handle("/app/", api_cfg.middlewareMetricsInc(app_handler))
	s_mux.HandleFunc("GET /admin/metrics", api_cfg.handlerMetricReqs)
	s_mux.HandleFunc("POST /admin/reset", api_cfg.handlerMetricReset)
	s_mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	s_mux.HandleFunc("GET /api/healthz", readinessHandler)

	server := http.Server{
		Addr:    ":" + port,
		Handler: s_mux,
	}
	server.ListenAndServe()
}

// handlers
func readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetricReqs(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
	<html>
		<body>
    		<h1>Welcome, Chirpy Admin</h1>
    		<p>Chirpy has been visited %d times!</p>
  		</body>
	</html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerMetricReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reset hits to 0"))
}

func validateChirpHandler(w http.ResponseWriter, req *http.Request) {
	type chirpBody struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	chirp_body := chirpBody{}
	err := decoder.Decode(&chirp_body)

	type errorStruct struct {
		Error string `json:"error"`
	}

	if err != nil {
		error_struct := errorStruct{
			Error: "Something went wrong",
		}
		data, err := json.Marshal(error_struct)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(data)
		return
	}

	if len(chirp_body.Body) > 140 {
		error_struct := errorStruct{
			Error: "Chirp is too long",
		}
		data, err := json.Marshal(error_struct)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(data)
		return
	}

	type validStruct struct {
		Valid bool `json:"valid"`
	}

	valid_struct := validStruct{
		Valid: true,
	}
	data, err := json.Marshal(valid_struct)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
