package main

import (
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
