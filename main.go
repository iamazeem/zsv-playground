package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html"
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
		peer := r.RemoteAddr

		log.Printf("[%v] %v", peer, *r)
		if err := r.ParseForm(); err != nil {
			log.Printf("[%v] failed to parse form, error: %v", peer, err)
			if _, err := w.Write([]byte(err.Error())); err != nil {
				log.Printf("[%v] failed to send 'parsing failed' response", peer)
			}
			return
		}

		version := r.FormValue("version")
		cli := r.FormValue("cli")
		csv := r.FormValue("csv")

		// log.Printf("[%v] version: [%v]", peer, version)
		// log.Printf("[%v] cli: [%v]", peer, cli)
		// log.Printf("[%v] csv: [%v]", peer, csv)

		zsv := getZsvExePath(version)
		cli = strings.Replace(cli, "zsv", zsv, 1)
		log.Printf("[%v] executing: [%v]", peer, cli)

		start := time.Now()

		cmd := exec.Command("sh", "-c", cli)
		cmd.Stdin = strings.NewReader(csv)
		output, err := cmd.CombinedOutput()

		elapsedTimeMsg := fmt.Sprintf("\n\n(elapsed time: %v)", time.Since(start))

		escapedOutput := html.EscapeString(string(output))
		if err != nil {
			log.Printf("[%v] failed to execute command, error: %v", peer, err)
			if !strings.HasSuffix(escapedOutput, "\n") {
				escapedOutput += "\n"
			}
			if _, err := w.Write([]byte(escapedOutput + err.Error() + elapsedTimeMsg)); err != nil {
				log.Printf("[%v] failed to send 'execution failed' response", peer)
			}
			return
		}

		// log.Printf("[%v] output: [%v]", peer, string(output))
		log.Printf("[%v] sending response [size: %v bytes]", peer, len(escapedOutput))
		if n, err := w.Write([]byte(escapedOutput + elapsedTimeMsg)); err != nil {
			log.Printf("[%v] failed to send response", peer)
		} else {
			log.Printf("[%v] sent response successfully [%v]", peer, n)
		}
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
