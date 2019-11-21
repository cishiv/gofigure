package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
     port := os.Getenv("PORT")
     log.Println("env port " + port)
     if port == "" {
     	defaultPort := "3000"
     	log.Println("no env var set for port, defaulting to " + port)
     	http.Handle("/", http.FileServer(http.Dir("./static")))
     	log.Println("starting server on port " + defaultPort)
     	http.ListenAndServe(":"+defaultPort, nil)
     } else {
		http.Handle("/", http.FileServer(http.Dir("./static")))
     	log.Println("starting server on port " + port)
     	log.Println("1 2 3")
     	http.ListenAndServe(":"+port, nil)
     }
   
}
