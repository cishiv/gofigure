package main

import (
	"log"
	"net/http"
)

func main() {
	 log.Println("starting server on port 3000")
	 http.Handle("/", http.FileServer(http.Dir("./static")))
     http.ListenAndServe(":3000", nil)
}
