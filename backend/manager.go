package main

import (
	"log"
	"github.com/prologic/bitcask"
	"crypto/sha256"
	"crypto/md5"
	"io/ioutil"
	"io"
	"bufio"
	"os"
	"os/exec"
	"strings"
	"encoding/hex"
	"time"
	"bytes"
	"strconv"
	"path/filepath"
)

 /**
 	Going to put all the helper methods and file hashing stuff here as part of an initial refactor to clean up the codebase
 */

 // In the grander scheme of things, global scope doesn't made sense, it's fine for now, but should be changed TODO
var registry []string
var buildHistory []Build
var actions []string
var whiteList []string

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

func Startup(db* bitcask.Bitcask) {
	// this should be parameterized, detected and injected TODO
	log.Println("building jobs")
	actions = append(actions, "./build build docker && kubectl delete --all deployments && kubectl delete --all services && kubectl apply -f deploy.yml")
	actions = append(actions, "./build build docker")
	actions = append(actions, "./build build bin")
	log.Println("building registry")
    // probably should always create the whitelist before the registry
    CreateWhiteList()
	BuildRegistry(db)

    DebugDB(db)
    // I probably need to think about this goroutine a bit
    // k8s is a bit heavy for every file change
    go DoEvery(2*time.Second, VerifyHashes, db , actions[2])

}

func BuildRegistry(db* bitcask.Bitcask) {
	log.Println("starting directory scan")
	RecursiveDirectoryCrawl(".", db)
	log.Println("computing hashes & creating Bitcask entires")
	for _, fn := range registry {
		hash := CalculateHash(fn)
		InsertRecord(fn, hash, db)
	}
}

// assume that a file called .fignore exists at ./fignore , if not return a warning to the terminal and possibly expose rest endpoint that frontend can poll to indicate that .fignore needs to be created
// we can use this `whitelist` file to ignore certain files in our hash recomputation so that we don't infinently kick off builds as result of builds
// this will probably self-alleviate when we start building in docker or we start producing artifacts in other directories. but for now we need this
// to avoid race conditions
// we can only ignore regular files for the moment and not directories (TODO - this weekend) we probably need regex matching for that
func CreateWhiteList () {
	file, err := os.Open("./.fignore")
	if err != nil {
		log.Println("no .fignore file found, race condition will ensue if jobs edit files -- will not create whitelist")

	} else {
		defer file.Close()
    	scanner := bufio.NewScanner(file)
	    for scanner.Scan() {
    		log.Println(scanner.Text())
        	whiteList = append(whiteList, scanner.Text())
    	}
    	if err := scanner.Err(); err != nil {
        	log.Fatal(err)
    	}
	}
}

func DebugDB(db* bitcask.Bitcask) {
	for _, fn := range registry {
		log.Println("from database: " + fn + ":" +   RetrieveHash(fn, db))
	}
}

func InsertRecord(absoluteFilePath string, hash string, db* bitcask.Bitcask) {
    db.Put([]byte(absoluteFilePath), []byte(hash))
}

func RetrieveHash(absoluteFilePath string, db* bitcask.Bitcask) string {
	val, _ := db.Get([]byte(absoluteFilePath))
	return CToGoString(val)
}

// we need to ignore .git
func RecursiveDirectoryCrawl(dirName string, db* bitcask.Bitcask) {
	files, err := ioutil.ReadDir(dirName)
	HandleError(err)
	for _, f := range files {
		fileOrDir, err := os.Stat(dirName + "/" + f.Name())
		HandleError(err)
		switch mode := fileOrDir.Mode(); {
		case mode.IsDir():
			// keep looking for files
			if !(f.Name() == ".git") {
				RecursiveDirectoryCrawl(dirName + "/" + f.Name(), db)
			}
		case mode.IsRegular():
			// O(n) brute force search in honour of silicon valley s06e04
			// if the file is whitelisted, don't add it to the registry
			toAdd := true
			for _, whitelisted := range whiteList {
				if (f.Name() == whitelisted) {
					toAdd = false
					log.Println(f.Name() + " is whitelisted, not adding to registry")
				}
			}
			if toAdd {
				absolutePath := dirName + "/" + f.Name()
				registry = append(registry, absolutePath)		
			} 
		}
	}
}

func CalculateHash(absoluteFilePath string) string {
	f, err := os.Open(absoluteFilePath)
  	HandleError(err)
  	defer f.Close()
 	h := sha256.New()
 	if _, err := io.Copy(h, f); err != nil {
    	log.Fatal(err)
  	}
  	return hex.EncodeToString(h.Sum(nil))
}

func HandleError(e error) {
	if e != nil {
		panic(e)
	}
}

func CompareHash(old string, new string) int {
	return strings.Compare(old, new)
}

func VerifyHashes(t time.Time, db *bitcask.Bitcask, action string) {
	for _, fn := range registry {
		oldHash := RetrieveHash(fn, db)
		log.Println(fn + ":" + oldHash)
		newHash := CalculateHash(fn)
		if !(CompareHash(oldHash, newHash) == 0) {
			InsertRecord(fn, newHash, db)
			log.Println(fn + " old hash" + oldHash + "new hash" + newHash + "changed detected - updating hash, action required")
			TakeAction(action)
		} 
	}
}

func TakeAction(action string) {
	log.Println("Taking action, running: "+action)
	startTime := GetTime()
	buildID := CalcMD5(startTime, action)
	build := Build{BuildID:buildID,Time:startTime,Action:action, Outcome:"started"}
	log.Println(build)
	buildHistory = append(buildHistory, build)
	cmd := exec.Command("/bin/sh", "-c", action)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	HandleError(err)
	log.Println(outb.String())
	log.Println(errb.String())
	log.Println("--------------------------------------------------------------------------------")
	b := &buildHistory[GetBuildByIDIndex(buildID)]
	// simulate a build time; just so we can observe the ui intergration
	// time.Sleep(10000 * time.Millisecond)
	b.Outcome = "success"
	log.Println(buildHistory)
	log.Println("control returned to gofigure")
}

func DoEvery(d time.Duration, f func(time.Time, *bitcask.Bitcask, string), db* bitcask.Bitcask, action string) {
	for x := range time.Tick(d) {
		f(x, db, action)
	}
}


func GetTime() string {
	// convert to ms
     currTime := time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
	 return strconv.FormatInt(currTime, 10)
}

func CalcMD5(buildStart string, action string) string {
	h := md5.New()
	io.WriteString(h, buildStart + " " + action)
	return hex.EncodeToString(h.Sum(nil))
}


/**
	Since we interact over REST the functions that the handlers invoke should live here (once we resolve global scope vars, we can move this into service.go)
**/

// TODO - Once we do vcs integration this will change
func GetBaseDir() string { 
 	ex, err := os.Executable()
    if err != nil {
        panic(err)
    }
    exPath := filepath.Dir(ex)
	return exPath
}

func ManualBuild(actionID string) {
	i, err := strconv.Atoi(actionID)
	HandleError(err)
	TakeAction(actions[i])
}

func GetActions() []string {
	return actions
}

func FetchHashes() []string {
	var hashes []string
	for _, fn := range registry {
		hashes = append(hashes, CalculateHash(fn) + " ------------> " + fn)
	}
	return hashes
}

func GetBuildHistory() []Build {
	return buildHistory
}