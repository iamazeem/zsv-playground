package main

import (
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	log.Printf("starting zsv playground [%v]", version)

	zsvVersions, err := setupCache()
	if err != nil {
		log.Fatalf("failed to set up zsv cache, %v", err)
	}

	zsvCLIsJson, ok := getCLIsJSON(zsvVersions)
	if !ok {
		log.Fatalf("failed to load CLI for all zsv versions")
	}

	startHTTPServer(zsvVersions, zsvCLIsJson)

	log.Print("exiting zsv playground, bye!")
}
