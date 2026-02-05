package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/MadridMAC/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits  atomic.Int32
	databaseQueries database.Queries
	userPlatform    string
}

func main() {
	godotenv.Load()

	const filepathRoot = "."
	const port = "8080"

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error opening DB_URL: %s", err)
	}
	dbQueries := database.New(db)

	api_cfg := apiConfig{
		fileserverHits:  atomic.Int32{},
		databaseQueries: *dbQueries,
		userPlatform:    platform,
	}

	s_mux := http.NewServeMux()
	app_handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	s_mux.Handle("/app/", api_cfg.middlewareMetricsInc(app_handler))
	s_mux.HandleFunc("GET /admin/metrics", api_cfg.handlerMetricReqs)
	// s_mux.HandleFunc("POST /admin/reset", api_cfg.handlerMetricReset)
	s_mux.HandleFunc("POST /admin/reset", api_cfg.deleteUsersHandler)
	//s_mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	s_mux.HandleFunc("POST /api/chirps", api_cfg.createChirpHandler)
	s_mux.HandleFunc("POST /api/users", api_cfg.createUserHandler)
	s_mux.HandleFunc("GET /api/healthz", readinessHandler)

	server := http.Server{
		Addr:    ":" + port,
		Handler: s_mux,
	}
	server.ListenAndServe()
}
