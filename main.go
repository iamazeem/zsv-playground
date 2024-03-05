package main

import (
	"html/template"
	"log"
	"net/http"
)

type data struct {
	PlaygroundVersion string
	ZsvVersions       []string
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	log.Printf("starting zsv playground")

	zsvVersions, err := setupZsvCache()
	if err != nil {
		log.Fatalf("failed to set up zsv cache, %v\n", err)
	}

	log.Printf("cached zsv versions: %v\n", zsvVersions)

	zsvExePaths := getZsvExePaths(zsvVersions)
	log.Printf("cached zsv binaries: %v\n", zsvExePaths)

	templates, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("failed to parse templates, %v\n", err)
	}

	d := data{
		PlaygroundVersion: version,
		ZsvVersions:       zsvVersions,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", d)
	})

	address := ":8080"
	log.Printf("starting http server [%v]\n", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalln(err)
	}
}
