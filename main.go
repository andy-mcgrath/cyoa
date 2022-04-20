package main

import (
	"encoding/json"
	"fmt"
	"html/template"
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
		Arc  string  `json:"arc"`
	}
)

const (
  pageTemplate = `
<!doctype html>
<html>
  <head>
    <title>{{.Title}}</title>
    <meta name="description" content="Our first page">
    <meta name="keywords" content="html tutorial template">
  </head>
  <body>
    <h1>{{.Title}}</h1>
    {{range .Story}}
    <p>.</p>
    <br>
    {{end}}
    {{range .Options}}
    <a href="/{{.Arc}}" >{{.Text}}</a>
    {{end}}
  </body>
</html>
`
)

func (p Page) print() {
  fmt.Printf("Title: %s\n\nStory:\n  ", p.Title)
  for _, v := range p.Story {
    fmt.Printf("%s\n", v)
  }
  fmt.Println()

  for _, v := range p.Options {
    fmt.Printf("%s ** %s **\n", v.Text ,v.Arc)
  }
}

func MapHandler(pages map[string]Page, tmpl *template.Template, fallback http.Handler) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    path := strings.Replace(r.URL.Path, "/", "", 1)
		if page, ok := pages[path]; ok {
			tmpl.Execute(w, page)
			return
		}

		fallback.ServeHTTP(w, r)
	})
}

func main() {
  pages := map[string]Page{}
  file := "gopher.json"

  data, err := os.ReadFile(file)
  if err != nil {
    panic(err)
  }
  err = json.Unmarshal(data, &pages)
  if err != nil {
    panic(err)
  }

  tmpl := template.Must(template.ParseFiles("page.html"))

  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
    http.Redirect(w, r, "/intro", http.StatusPermanentRedirect)
  })

  storyHandler := MapHandler(pages, tmpl, mux)

  fmt.Println("Web Server running on http://localhost:8080/intro")
  http.ListenAndServe(":8080", storyHandler)
}
