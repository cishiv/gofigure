package main

import (
	"log"
	"github.com/prologic/bitcask"
	"crypto/sha256"
	"io/ioutil"
	"io"
	"os"
	"strings"
	"encoding/hex"
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
var registry []string
func main() {
	// init db
	 db, _ := bitcask.Open("/tmp/db")
     defer db.Close()
     startup(db)
}

// we need to convert unfortunately :(
func CToGoString(c []byte) string {
    n := -1
    for i, b := range c {
        if b == 0 {
            break
        }
        n = i
    }
    return string(c[:n+1])
}

func startup(db* bitcask.Bitcask) {
	log.Println("building registry")
	buildRegistry(db)
}

func buildRegistry(db* bitcask.Bitcask) {
	log.Println("starting directory scan")
	recursiveDirectoryCrawl(".", db)
	log.Println("computing hashes")
	for _, fn := range registry {
		log.Println(fn + " " + calculateHash(fn))
	}
	log.Println("creating entries")
}



func insertRecord(absoluteFilePath string, hash string, db* bitcask.Bitcask) {
    db.Put([]byte(absoluteFilePath), []byte(hash))
}

func retrieveHash(absoluteFilePath string, db* bitcask.Bitcask) string {
	val, _ := db.Get([]byte(absoluteFilePath))
	return CToGoString(val)
}

func recursiveDirectoryCrawl(dirName string, db* bitcask.Bitcask) {
	files, err := ioutil.ReadDir(dirName)
	handleError(err)
	for _, f := range files {
		fileOrDir, err := os.Stat(dirName + "/" + f.Name())
		handleError(err)
		switch mode := fileOrDir.Mode(); {
		case mode.IsDir():
			// keep looking for files
			recursiveDirectoryCrawl(dirName + "/" + f.Name(), db)
		case mode.IsRegular():
			absolutePath := dirName + "/" + f.Name();
			registry = append(registry, absolutePath)
		}
	}
}

func calculateHash(absoluteFilePath string) string {
	f, err := os.Open(absoluteFilePath)
  	handleError(err)
  	defer f.Close()
 	h := sha256.New()
 	if _, err := io.Copy(h, f); err != nil {
    	log.Fatal(err)
  	}
  	return hex.EncodeToString(h.Sum(nil))
}
func handleError(e error) {
	if e != nil {
		panic(e)
	}
}

func compareHash(old string, new string) int {
	return strings.Compare(old, new)
}

func recomputeHash() {}