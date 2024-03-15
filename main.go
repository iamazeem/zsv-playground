package main

import (
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"
)

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
	log.Printf("starting zsv playground [%v]", version)

	zsvVersions, err := setupZsvCache()
	if err != nil {
		log.Fatalf("failed to set up zsv cache, %v", err)
	}

	log.Printf("cached zsv versions: %v", zsvVersions)

	zsvExePaths := getZsvExePaths(zsvVersions)
	log.Printf("cached zsv binaries: %v", zsvExePaths)

	clis, ok := loadCLIs(zsvVersions)
	if !ok {
		log.Fatalf("failed to load CLI for all zsv versions")
	}

	clisJson, err := json.Marshal(clis)
	if err != nil {
		log.Fatalf("failed to marshal CLIs to JSON, error: %v", err)
	}

	clisJsonStr := string(clisJson)
	// log.Print(clisJsonStr)

	templates, err := template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		log.Fatalf("failed to parse templates, %v", err)
	}

	data := struct {
		PlaygroundVersion string
		ZsvVersions       []string
		ZsvCLIsJson       string
	}{
		PlaygroundVersion: version,
		ZsvVersions:       zsvVersions,
		ZsvCLIsJson:       clisJsonStr,
	}

	http.Handle("/static/", http.FileServer(http.FS(staticFS)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			log.Print(err)
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		log.Print(*r)
		if err := r.ParseForm(); err != nil {
			log.Printf("failed to parse form, error: %v", err)
			w.Write([]byte("Failure!"))
			return
		}

		version := r.FormValue("version")
		cli := r.FormValue("cli")
		csv := r.FormValue("csv")

		log.Printf("version: [%v]", version)
		log.Printf("cli: [%v]", cli)
		log.Printf("csv: [%v]", csv)

		zsv := getZsvExePath(version)
		cli = strings.Replace(cli, "zsv", zsv, 1)
		log.Printf("executing: %v", cli)

		cmd := exec.Command("sh", "-c", cli)
		cmd.Stdin = strings.NewReader(csv)
		output, err := cmd.Output()
		if err != nil {
			log.Printf("failed to execute command, error: %v", err)
			w.Write([]byte(err.Error()))
			return
		}
		log.Printf("output: [%v]", string(output))
		w.Write(output)
	})

	// start http server and wait for SIGINT

	address := ":8080"
	server := &http.Server{Addr: address}
	go func() {
		log.Printf("starting http server [graceful shutdown on SIGINT]")
		if err := server.ListenAndServe(); err != nil {
			log.Print(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown, error: %v", err)
	}

	log.Print("exiting zsv playground, bye!")
}
