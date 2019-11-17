package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to GoFigure")
}

func GetHashInfo(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	count := FetchHash("./README.md")
	if err := json.NewEncoder(w).Encode(count); err != nil {
		panic(err)
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}