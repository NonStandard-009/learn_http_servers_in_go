package main

import (
	"fmt"
	"net/http"
)

func main() {
	newMultiplexer := http.NewServeMux()

	newServer := http.Server{
		Handler: newMultiplexer,
		Addr:    ":8080",
	}

	if err := newServer.ListenAndServe(); err != nil {
		fmt.Printf("%v", err)
	}
}
