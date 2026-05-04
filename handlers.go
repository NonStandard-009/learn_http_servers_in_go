package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/NonStandard-009/chirpy/internal/auth"
	"github.com/NonStandard-009/chirpy/internal/database"
	"github.com/google/uuid"
)

var profanities = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

const defaultTokenExpiration = time.Hour

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	jwtSecret      string
}

func (cfg *apiConfig) createUsersHandler(w http.ResponseWriter, r *http.Request) {
	tmpForDecoding := UserReqParams{}
	if err := helperDecode(r, &tmpForDecoding); err != nil {
		respondWithError(w, http.StatusBadRequest, "Error while trying to decode request")
		return
	}

	hashedPwd, err := auth.HashPassword(tmpForDecoding.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error while trying to hash password")
		return
	}

	newUserParams := database.CreateUserParams{
		Email:          tmpForDecoding.Email,
		HashedPassword: hashedPwd,
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), newUserParams)
	if err != nil {
		log.Printf("Error while trying to create user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error while trying to create user")
		return
	}

	newUser := UserJSON{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, newUser)
}

func (cfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	user := UserReqParams{}

	if err := helperDecode(r, &user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Error while trying to decode request")
		return
	}

	tokenExpiration := time.Duration(user.ExpiresInSeconds) * time.Second
	if tokenExpiration > defaultTokenExpiration || tokenExpiration <= 0 {
		tokenExpiration = defaultTokenExpiration
	}

	dbUser, err := cfg.dbQueries.GetUserByEmail(r.Context(), user.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	confirm, err := auth.CheckPassword(user.Password, dbUser.HashedPassword)
	if err != nil {
		log.Printf("Error while trying to confirm password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error while trying to confirm password")
		return
	}

	if !confirm {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	newToken, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, tokenExpiration)
	if err != nil {
		log.Printf("Error while trying to create user token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error while trying to create user token")
		return
	}

	respondUser := UserJSON{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     newToken,
	}

	respondWithJSON(w, http.StatusOK, respondUser)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	reqToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("GetBearerToken failed: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(reqToken, cfg.jwtSecret)
	if err != nil {
		log.Printf("ValidateJWT failed: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpBody := ChirpRequestJSON{}

	if err := helperDecode(r, &chirpBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Error while trying to decode request")
		return
	}

	if len(chirpBody.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanMsg := cleanResponse(profanities, chirpBody.Body)

	newChirpParams := database.CreateChirpParams{
		Body:   cleanMsg,
		UserID: userID,
	}

	newChirp, err := cfg.dbQueries.CreateChirp(r.Context(), newChirpParams)
	if err != nil {
		log.Printf("Error while trying to create chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error while trying to create chirp")
		return
	}

	responseChirp := ChirpJSON{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, responseChirp)
}

func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	type ChirpJSONArray struct {
		Chirps []ChirpJSON `json:"chirps"`
	}

	getChirps, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error while trying to retrieve ALL chirps: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error while trying to retrieve ALL chirps")
		return
	}

	listChirpsForResponse := make([]ChirpJSON, len(getChirps))

	for i, chirp := range getChirps {
		chirpToJSON := ChirpJSON{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		listChirpsForResponse[i] = chirpToJSON
	}

	respondWithJSON(w, http.StatusOK, listChirpsForResponse)
}

func (cfg *apiConfig) getSingleChirpHandler(w http.ResponseWriter, r *http.Request) {
	pattern := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(pattern)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error while trying to get ID from URL")
		return
	}

	getChirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}

	chirp := ChirpJSON{
		ID:        getChirp.ID,
		CreatedAt: getChirp.CreatedAt,
		UpdatedAt: getChirp.UpdatedAt,
		Body:      getChirp.Body,
		UserID:    getChirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirp)
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
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err := cfg.dbQueries.DeleteAllUsers(r.Context()); err != nil {
		log.Printf("Error while trying to delete users: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error while trying to delete users")
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
