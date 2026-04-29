package main

import (
	"log"
	"net/http"
)

const projectRoot = "/home/thanatos/Workspace/boot_dev/back_end_dev_path/learn_http_servers_in_go/"
const port = "8080"

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(projectRoot)))

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from '%s' on port: %s\n", projectRoot, port)
	log.Fatal(server.ListenAndServe())
}
