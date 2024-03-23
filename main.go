package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	versionFlag := flag.Bool("version", false, "show version")
	addressFlag := flag.String("address", "0.0.0.0:8080", "host:port for HTTP server")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("zsv-playground %v\n", version)
		return
	}

	log.Printf("starting zsv-playground %v [os: %v, arch: %v]", version, runtime.GOOS, runtime.GOARCH)

	zsvVersions, err := setupCache()
	if err != nil {
		log.Fatalf("failed to set up zsv cache, %v", err)
	}

	zsvCLIsJson, ok := getCLIsJSON(zsvVersions)
	if !ok {
		log.Fatalf("failed to load CLI for all zsv versions")
	}

	startHTTPServer(*addressFlag, zsvVersions, zsvCLIsJson)

	log.Print("exiting zsv playground, bye!")
}
