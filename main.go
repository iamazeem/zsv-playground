package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

// type Page struct {
// 	Title string
// 	Body  []byte
// }

// func (p *Page) save() error {
// 	filename := p.Title + ".txt"
// 	return os.WriteFile(filename, p.Body, 0600)
// }

// func loadPage(title string) (*Page, error) {
// 	filename := title + ".txt"
// 	body, err := os.ReadFile(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Page{Title: title, Body: body}, nil
// }

// func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
// 	p, err := loadPage(title)
// 	if err != nil {
// 		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
// 		return
// 	}
// 	renderTemplate(w, "view", p)
// }

// func editHandler(w http.ResponseWriter, r *http.Request, title string) {
// 	p, err := loadPage(title)
// 	if err != nil {
// 		p = &Page{Title: title}
// 	}
// 	renderTemplate(w, "edit", p)
// }

// func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
// 	body := r.FormValue("body")
// 	p := &Page{Title: title, Body: []byte(body)}
// 	err := p.save()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	http.Redirect(w, r, "/view/"+title, http.StatusFound)
// }

var templates = template.Must(template.ParseFiles("templates/index.html"))

// func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
// 	err := templates.ExecuteTemplate(w, tmpl+".html", p)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	}
// }

// var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		m := validPath.FindStringSubmatch(r.URL.Path)
// 		if m == nil {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		fn(w, r, m[2])
// 	}
// }

func generateZsvExePath(version string) string {
	return fmt.Sprintf("%v/%v/%v/bin/zsv", cacheDir, version, triplet)
}

func getZsvExePaths(versions []string) []string {
	bins := []string{}
	for _, v := range versions {
		zsv := generateZsvExePath(v)
		bins = append(bins, zsv)
	}
	return bins
}

func main() {
	zsvVersions, err := setupZsvCache()
	if err != nil {
		fmt.Printf("failed to set up zsv cache, %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("cached zsv versions: %v\n", zsvVersions)

	zsvExePaths := getZsvExePaths(zsvVersions)
	fmt.Printf("cached zsv binaries: %v\n", zsvExePaths)

	// http.HandleFunc("/view/", makeHandler(viewHandler))
	// http.HandleFunc("/edit/", makeHandler(editHandler))
	// http.HandleFunc("/save/", makeHandler(saveHandler))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	})

	// err := http.ListenAndServe(":8080", nil)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
}
