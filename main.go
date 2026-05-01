package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/NonStandard-009/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const port = "8080"

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error opening database")
	}

	dbQueries := database.New(db)
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	cfg := apiConfig{
		dbQueries: dbQueries,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validedChirpHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from '/' on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
