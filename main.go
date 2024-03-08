package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

type data struct {
	PlaygroundVersion string
	ZsvVersions       []string
}

var (
	//go:embed templates/index.html
	templatesFS embed.FS

	//go:embed static/bootstrap.min@v5.3.3.css static/htmx.org@1.9.10.js
	staticFS embed.FS
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	log.Printf("starting zsv playground")

	zsvVersions, err := setupZsvCache()
	if err != nil {
		log.Fatalf("failed to set up zsv cache, %v", err)
	}

	log.Printf("cached zsv versions: %v", zsvVersions)

	zsvExePaths := getZsvExePaths(zsvVersions)
	log.Printf("cached zsv binaries: %v", zsvExePaths)

	zsvCLI, ok := loadCLIForAllZsvVersions(zsvVersions)
	if !ok {
		log.Fatalf("failed to load CLI for all zsv versions")
	}

	for version, subcommands := range zsvCLI {
		log.Printf("listing all commands with flags for %v", version)
		log.Print(subcommands)
	}

	templates, err := template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		log.Fatalf("failed to parse templates, %v", err)
	}

	d := data{
		PlaygroundVersion: version,
		ZsvVersions:       zsvVersions,
	}

	http.Handle("/static/", http.FileServer(http.FS(staticFS)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", d)
	})

	address := ":8080"
	log.Printf("starting http server [%v]", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalln(err)
	}
}
