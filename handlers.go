package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MadridMAC/chirpy/internal/database"
	"github.com/google/uuid"
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

/*
func (cfg *apiConfig) handlerMetricReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reset hits to 0"))
}
*/

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, req *http.Request) {
	type chirpBody struct {
		Body    string    `json:"body"`
		User_Id uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	chirp_body := chirpBody{}
	err := decoder.Decode(&chirp_body)

	// respond w/ error if something goes wrong decoding chirpBody
	if err != nil {
		respondWithError(w, 400, "Something went wrong")
		return
	}

	// respond w/ 'too long' if Body has more than 140 characters
	if len(chirp_body.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	// legacy profanityChecker; unsure if still needed
	// cleanedBody := profanityChecker(chirp_body.Body)

	// debug info
	// fmt.Printf("Chirp Body: %v\n", chirp_body.Body)
	// fmt.Printf("Chirp UserID: %v\n", chirp_body.User_Id)

	// if chirp is valid, save to DB
	chirpParams := database.CreateChirpParams{
		Body:   chirp_body.Body,
		UserID: chirp_body.User_Id,
	}
	chirp, err := cfg.databaseQueries.CreateChirp(req.Context(), chirpParams)
	if err != nil {
		respondWithError(w, 400, "Error occurred during Chirp creation: "+err.Error())
		return
	}

	// create outputs truct for respondWithJSON
	type responseStruct struct {
		Id         uuid.UUID `json:"id"`
		Created_At time.Time `json:"created_at"`
		Updated_At time.Time `json:"updated_at"`
		Body       string    `json:"body"`
		User_Id    uuid.UUID `json:"user_id"`
	}

	output := responseStruct{
		Id:         chirp.ID,
		Created_At: chirp.CreatedAt,
		Updated_At: chirp.UpdatedAt,
		Body:       chirp.Body,
		User_Id:    chirp.UserID,
	}

	// when successful, respond with HTTP 201 and full chirp resource
	respondWithJSON(w, 201, output)

}

// createUserHandler
func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, req *http.Request) {
	// create request struct and decode
	type emailBody struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(req.Body)
	email_body := emailBody{}
	err := decoder.Decode(&email_body)

	// respond w/ error if something goes wrong decoding emailBody
	if err != nil {
		respondWithError(w, 400, "Something went wrong while decoding email request")
	}

	// create a new user via CreateUser and email_body.Email
	user, err := cfg.databaseQueries.CreateUser(req.Context(), email_body.Email)
	if err != nil {
		respondWithError(w, 400, "Error occurred during user creation")
		return
	}

	// create response struct and build output
	type responseStruct struct {
		Id         uuid.UUID `json:"id"`
		Created_At time.Time `json:"created_at"`
		Updated_At time.Time `json:"updated_at"`
		Email      string    `json:"email"`
	}

	output := responseStruct{
		Id:         user.ID,
		Created_At: user.CreatedAt,
		Updated_At: user.UpdatedAt,
		Email:      email_body.Email,
	}

	// respond with proper output and HTTP code 201 Created
	respondWithJSON(w, 201, output)
}

func (cfg *apiConfig) deleteUsersHandler(w http.ResponseWriter, req *http.Request) {
	if cfg.userPlatform != "dev" {
		respondWithError(w, 403, "403 Forbidden")
		return
	}

	err := cfg.databaseQueries.DeleteUsers(req.Context())
	if err != nil {
		respondWithError(w, 500, "Error deleting users from database")
	}
}

/*
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
*/
