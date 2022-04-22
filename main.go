package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type (
	Page struct {
		Title   string   `json:"title"`
		Story   []string `json:"story"`
		Options []Option `json:"options"`
	}
	Option struct {
		Text string `json:"text"`
		Arc  string `json:"arc"`
	}
)

func mapHandler(pages map[string]Page, tmpl *template.Template, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.Replace(r.URL.Path, "/", "", 1)

		if page, ok := pages[path]; ok {
			tmpl.Execute(w, page)
			return
		}

		fallback.ServeHTTP(w, r)
		return
	})
}

func loadStory(filePath string) (pages map[string]Page, err error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &pages)
	if err != nil {
		return
	}
	return
}

func defaultMux() *http.ServeMux {
	assetsFs := http.FileServer(http.Dir("web/assets"))
	imagesFs := http.FileServer(http.Dir("web/images"))
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", assetsFs))
	mux.Handle("/images/", http.StripPrefix("/images/", imagesFs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/intro", http.StatusPermanentRedirect)
			return
		}
		http.NotFound(w, r)
	})
	return mux
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
			log.Fatal("$PORT must be set")
	}

	storyFile := "web/stories/gopher.json"
	tmplFile := "web/template/index.gotmpl"

	pages, err := loadStory(storyFile)
	if err != nil {
		panic(err)
	}

	tmpl := template.Must(template.ParseFiles(tmplFile))

	mux := defaultMux()

	storyHandler := mapHandler(pages, tmpl, mux)

	fmt.Printf("Web Server running on http://localhost:%s/\n", port)
	http.ListenAndServe(":" + port, storyHandler)
}
