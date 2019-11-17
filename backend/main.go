package main

import (
	"log"
	"github.com/prologic/bitcask"
	"crypto/sha256"
	"crypto/md5"
	"io/ioutil"
	"io"
	"os"
	"os/exec"
	"strings"
	"encoding/hex"
	"time"
	"bytes"
	"strconv"
	"net/http"
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
var buildHistory []Build
var actions []string

type Build struct {
	BuildID string `json:"buildID"`
	Time string `json:"time"`
	Action string `json:"action"`
	Outcome string `json:"outcome"`
}

func getBuildByIDIndex(buildID string) int {
	for i,b := range buildHistory {
		if(b.BuildID == buildID) {
			return i
		}
	}
	return -1
}
func (build Build) setOutcome(result string) {
	build.Outcome = result;
}

func main() {
	// init db
	 actions = append(actions, "./build build docker && kubectl delete --all deployments && kubectl delete --all services && kubectl apply -f deploy.yml")
	 actions = append(actions, "./build build docker")
	 actions = append(actions, "go build")
	 db, _ := bitcask.Open("/tmp/db")
     defer db.Close()
     startup(db)
     debugDB(db)
     http.ListenAndServe(":3000", nil)
     // I probably need to think about this goroutine a bit
     go doEvery(2*time.Second, verifyHashes, db , actions[0])
     router := NewRouter()
	 log.Fatal(http.ListenAndServe(":8084", router))
     
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
	log.Println("computing hashes & creating Bitcask entires")
	for _, fn := range registry {
		hash := calculateHash(fn)
		log.Println(fn + " " + hash)
		insertRecord(fn, hash, db)
	}
}

func debugDB(db* bitcask.Bitcask) {
	for _, fn := range registry {
		log.Println("from database: " + fn + ":" +   retrieveHash(fn, db))
	}
}
func insertRecord(absoluteFilePath string, hash string, db* bitcask.Bitcask) {
    db.Put([]byte(absoluteFilePath), []byte(hash))
}

func retrieveHash(absoluteFilePath string, db* bitcask.Bitcask) string {
	val, _ := db.Get([]byte(absoluteFilePath))
	return CToGoString(val)
}

// we need to ignore .git
func recursiveDirectoryCrawl(dirName string, db* bitcask.Bitcask) {
	files, err := ioutil.ReadDir(dirName)
	handleError(err)
	for _, f := range files {
		fileOrDir, err := os.Stat(dirName + "/" + f.Name())
		handleError(err)
		switch mode := fileOrDir.Mode(); {
		case mode.IsDir():
			// keep looking for files
			if !(f.Name() == ".git") {
				recursiveDirectoryCrawl(dirName + "/" + f.Name(), db)
			}
		case mode.IsRegular():
			absolutePath := dirName + "/" + f.Name()
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

func verifyHashes(t time.Time, db *bitcask.Bitcask, action string) {
	for _, fn := range registry {
		oldHash := retrieveHash(fn, db)
		newHash := calculateHash(fn)
		if !(compareHash(oldHash, newHash) == 0) {
			insertRecord(fn, newHash, db)
			log.Println("changed detected - updating hash, action required")
			takeAction(action)
		} 
	}
}

func takeAction(action string) {
	log.Println("Taking action, running: "+action)
	startTime := getTime()
	buildID := calcMD5(startTime, action)
	build := Build{BuildID:buildID,Time:startTime,Action:action, Outcome:"started"}
	log.Println(build)
	buildHistory = append(buildHistory, build)
	cmd := exec.Command("/bin/sh", "-c", action)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	handleError(err)
	log.Println(outb.String())
	log.Println(errb.String())
	log.Println("--------------------------------------------------------------------------------")
	b := &buildHistory[getBuildByIDIndex(buildID)]
	// simulate a build time; just so we can observe the ui intergration
	// time.Sleep(10000 * time.Millisecond)
	b.Outcome = "success"
	log.Println(buildHistory)
	log.Println("control returned to gofigure")
}

func doEvery(d time.Duration, f func(time.Time, *bitcask.Bitcask, string), db* bitcask.Bitcask, action string) {
	for x := range time.Tick(d) {
		f(x, db, action)
	}
}

func FetchHashes() []string {
	var hashes []string
	for _, fn := range registry {
		hashes = append(hashes, calculateHash(fn) + " ------------> " + fn)
	}
	return hashes
}

func GetBuildHistory() []Build {
	return buildHistory
}

func getTime() string {
	// convert to ms
     currTime := time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
	 return strconv.FormatInt(currTime, 10)
}

func calcMD5(buildStart string, action string) string {
	h := md5.New()
	io.WriteString(h, buildStart + " " + action)
	return hex.EncodeToString(h.Sum(nil))
}

func GetBaseDir() string {
	return "github.com/cishiv/gofigure/"
}

func ManualBuild(actionID string) {
	i, err := strconv.Atoi(actionID)
	handleError(err)
	takeAction(actions[i])
}

func GetActions() []string {
	return actions
}