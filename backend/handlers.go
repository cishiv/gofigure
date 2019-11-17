package main

import (
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"github.com/gorilla/mux"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to GoFigure")
}

// recalculate and display all hashes
func GetHashInfo(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	hashes := FetchHashes()
	log.Println("returning hash to upstream")
	if err := json.NewEncoder(w).Encode(hashes); err != nil {
		panic(err)
	}
}

func GetJobs(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	jobs := GetActions()
	log.Println(jobs)
	log.Println("returning jobs to upstream")
	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		panic(err)
	}

}

func GetHistory(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	builds := GetBuildHistory()
	log.Println(builds)
	log.Println("returning build history to upstream")
	if err := json.NewEncoder(w).Encode(builds); err != nil {
		panic(err)
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func GetProjectName(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	projectName := GetBaseDir()
	if err := json.NewEncoder(w).Encode(projectName); err != nil {
		panic(err)
	}
}

func StartBuild(w http.ResponseWriter, r * http.Request) {
	enableCors(&w)
	id := ExtractVar(r, "id")
	ManualBuild(id)
}

// testing the new active build prompt
func ExtractVar(r *http.Request, varName string) string {
    	vars := mux.Vars(r)
     	return vars[varName]
}