package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const projectRoot = "/home/thanatos/Workspace/boot_dev/back_end_dev_path/learn_http_servers_in_go/"
const port = "8080"

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Hits: %d\n", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "Resetting hit counter\n")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Printf("Error writting response: %v\n", err)
	}
}

func main() {
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(projectRoot)))
	var cfg apiConfig

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(handler))

	mux.HandleFunc("/healthz", serverHandler)
	mux.HandleFunc("/metrics", cfg.metricsHandler)
	mux.HandleFunc("/reset", cfg.resetHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from '%s' on port: %s\n", projectRoot, port)
	log.Fatal(server.ListenAndServe())
}
