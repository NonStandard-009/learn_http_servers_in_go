package main

import (
	"log"
	"net/http"
)

const projectRoot = "/home/thanatos/Workspace/boot_dev/back_end_dev_path/learn_http_servers_in_go/"
const port = "8080"

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

	mux := http.NewServeMux()
	mux.Handle("/app/", handler)
	mux.HandleFunc("/healthz", serverHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from '%s' on port: %s\n", projectRoot, port)
	log.Fatal(server.ListenAndServe())
}
