package main

import (
	"log"
	"net/http"
)

const port = "8080"

func main() {
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	var cfg apiConfig

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /healthz", healthzHandler)
	mux.HandleFunc("GET /metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /reset", cfg.resetHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from '/' on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
