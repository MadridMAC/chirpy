package main

import (
	"fmt"
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
	s_mux.HandleFunc("/metrics/", api_cfg.handlerMetricReqs)
	s_mux.HandleFunc("/reset/", api_cfg.handlerMetricReset)
	s_mux.HandleFunc("/healthz", readinessHandler)

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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerMetricReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reset hits to 0"))
}
