package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type Error struct {
		Error string `json:"error"`
	}

	respondWithJSON(w, code, Error{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling data: %v", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func cleanResponse(profanities map[string]struct{}, body string) string {
	wordsInMsg := strings.Split(body, " ")

	for i, word := range wordsInMsg {
		loweredWord := strings.ToLower(word)
		if _, ok := profanities[loweredWord]; ok {
			wordsInMsg[i] = "****"
		}
	}

	return strings.Join(wordsInMsg, " ")
}
