package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type (
	Page struct {
		Parent  string   `json:"parent,omitempty"`
		Title   string   `json:"title"`
		Story   []string `json:"story"`
		Options []Option `json:"options"`
	}
	Option struct {
		Text string `json:"text"`
		Arc  string `json:"arc"`
	}
)

var (
	stories []string
	pages map[string]Page
)

func isStory(ctx context.Context, rdb *redis.Client, story string) bool {
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second * 5)
	defer cancel()

	s, err := rdb.SMIsMember(ctxTimeout, "stories", story).Result()
	if err != nil {
		fmt.Printf("redis SMIsMember 'stories', '%s' failed: %s\n", story, err.Error())
		return false
	}

	return s[0]
}

func urlPathToStoryPage(url string) (story, page string) {
	urlEl := strings.Split(strings.ToLower(url), "/")
	switch len(urlEl) {
	case 2:
		story = urlEl[1]
		page = "intro"
	case 3:
		story = urlEl[1]
		page = urlEl[2]
	default:
		story = ""
		page = ""
	}
	return story, page
}

func redisMapHandler(ctx context.Context, rdb *redis.Client, tmpl *template.Template, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		story, page := urlPathToStoryPage(r.URL.Path)

		if story == "" || page == "" || !isStory(ctx, rdb, story) {
			fallback.ServeHTTP(w, r)
			return
		}


		ctxTimeout, cancel := context.WithTimeout(ctx, time.Second * 5)
		defer cancel()

		key := fmt.Sprintf("%s:%s", story, page)

		rGet, err := rdb.Get(ctxTimeout, key).Result()
		if err != nil {
			fmt.Printf("redis get '%s' failed: %s\n", key, err.Error())
			fallback.ServeHTTP(w, r)
			return
		}

		p := Page{}
		err = json.Unmarshal([]byte(rGet), &p)
		if err != nil {
			fmt.Printf("json unmarshal failed: %s", err.Error())
			http.NotFound(w, r)
			return
		}

		p.Parent = story

		err = tmpl.Execute(w, p)
		if err != nil {
			fmt.Println(err.Error())
		}

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
			http.Redirect(w, r, "/gopher", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	return mux
}

func main() {
	ctx := context.Background()
	rdb, err := newRedisClient(ctx)
	if err != nil {
		log.Fatalf("redis client error: %s\n\n", err.Error())
	}
	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	tmplFile := "web/template/index.gotmpl"

	tmpl := template.Must(template.ParseFiles(tmplFile))

	mux := defaultMux()

	storyHandler := redisMapHandler(ctx, rdb, tmpl, mux)

	fmt.Printf("Web Server running on http://localhost:%s/\n", port)
	http.ListenAndServe(":" + port, storyHandler)
}


func newRedisClient(ctx context.Context) (*redis.Client, error) {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return nil, fmt.Errorf("redis parse url failed: %w",err)
	}

	c := redis.NewClient(opt)

	return c, nil
}