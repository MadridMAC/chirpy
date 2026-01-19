package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

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

	// respond w/ error if something goes wrong decoding chirpBody
	if err != nil {
		respondWithError(w, 400, "Something went wrong")
	}

	// respond w/ 'too long' if Body has more than 140 characters
	if len(chirp_body.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
	}

	// clean profane words from input using profanityChecker, then send 200 JSON response
	cleanedBody := profanityChecker(chirp_body.Body)
	type cleanStruct struct {
		Cleaned_Body string `json:"cleaned_body"`
	}
	cleaned_struct := cleanStruct{
		Cleaned_Body: cleanedBody,
	}
	respondWithJSON(w, 200, cleaned_struct)
}

func profanityChecker(body string) string {
	profanity_list := []string{"kerfuffle", "sharbert", "fornax"}
	splitAndLowered := strings.Split(body, " ")
	for i, word := range splitAndLowered {
		if slices.Contains(profanity_list, strings.ToLower(word)) {
			splitAndLowered[i] = "****"
		}
	}
	cleanedWord := strings.Join(splitAndLowered, " ")
	return cleanedWord
}
