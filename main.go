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
)

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
		log.Fatalf("failed to parse templates, %v\n", err)
	}

	d := data{
		PlaygroundVersion: version,
		ZsvVersions:       zsvVersions,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", d)
	})

	http.HandleFunc("/zsv/commands", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", d)
	})

	http.HandleFunc("/zsv/flags", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", d)
	})

	// address := ":8080"
	// log.Printf("starting http server [%v]\n", address)
	// if err := http.ListenAndServe(address, nil); err != nil {
	// 	log.Fatalln(err)
	// }
}
