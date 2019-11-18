package main

import (
	"log"
	"net/http"
	"github.com/prologic/bitcask"
)

/**
	My personal CI tool that runs my build scripts and propogates my diffs into minikube
	Basic lifecycle:
	- startup
		- recursively build a registry of files and sha256 hashes
		NOTE: we do not compute hashes on directories
	- monitor
		- every 30s recursively rebuild hashes and compare. 
		- if changed hashes, run:
			- build bin
			- build docker
			- kubectl delete --all deployments
			- kubectl delete --all services
			- kubectl apply -f deploy.yml

	Currently using github.com/prologic/bitcask as an in memory k:v store for the file registry

*/
// FIXME!!!! STARTING BUILDS CAUSES MORE BUILDS
// SAFEGUARD IT


func main() {
	db, _ := bitcask.Open("/tmp/db")
    defer db.Close()
	Startup(db)
    router := NewRouter()
	log.Fatal(http.ListenAndServe(":8084", router))
     
}

// we need to convert unfortunately :(


