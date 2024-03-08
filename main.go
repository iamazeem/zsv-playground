package main

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
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

	// start http server and wait for SIGINT

	address := ":8080"
	server := &http.Server{Addr: address}
	go func() {
		log.Printf("starting http server")
		if err := server.ListenAndServe(); err != nil {
			log.Print(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Print("waiting for SIGINT to shutdown")
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Print("exiting zsv playground, bye!")
}
