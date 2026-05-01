package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/NonStandard-009/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (cfg *apiConfig) createUsersHandler(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email string `json:"email"`
	}
	newUserParams := params{}

	if err := helperDecode(r, &newUserParams); err != nil {
		respondWithError(w, 400, "Error while trying to decode request")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), newUserParams.Email)
	if err != nil {
		log.Printf("Error while trying to create user: %v", err)
		respondWithError(w, 500, "Error while trying to create user")
		return
	}

	newUser := UserMain{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, 201, newUser)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	newChirp := ChirpMain{}

	if err := helperDecode(r, &newChirp); err != nil {
		respondWithError(w, 400, "Error while trying to decode request")
		return
	}

	if len(newChirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	cleanMsg := cleanResponse(
		map[string]struct{}{
			"kerfuffle": {},
			"sharbert":  {},
			"fornax":    {},
		}, newChirp.Body)

	newChirpParams := database.CreateChirpParams{
		Body:   cleanMsg,
		UserID: newChirp.UserID,
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), newChirpParams)
	if err != nil {
		log.Printf("Error while trying to create chirp: %v", err)
		respondWithError(w, 500, "Error while trying to create chirp")
		return
	}

	newChirp = ChirpMain{
		Body:   chirp.Body,
		UserID: chirp.UserID,
	}

	respondWithJSON(w, 201, newChirp)
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	htmlContent := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`,
		cfg.fileserverHits.Load())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte(htmlContent))
	if err != nil {
		log.Printf("Error writting response: %v\n", err)
	}
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Forbidden")
		return
	}

	if err := cfg.dbQueries.DeleteAllUsers(r.Context()); err != nil {
		log.Printf("Error while trying to delete users: %v", err)
		respondWithError(w, 500, "Error while trying to delete users")
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "Resetting hit counter and deleting ALL users\n")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Printf("Error writting response: %v\n", err)
	}
}

func helperDecode(r *http.Request, payload any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(payload); err != nil {
		return fmt.Errorf("Failure to decode request: %w", err)
	}
	return nil
}
