package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	blade "github.com/zane-xin/sparq-blade"
)

func main() {
	renderer := blade.New(
		blade.WithFS(blade.OSFS{Root: viewsRoot()}),
		blade.WithDirective("badge", func(expression string) (string, error) {
			return fmt.Sprintf(`<strong class="badge">{{ bladeEcho . %q }}</strong>`, expression), nil
		}),
		blade.WithBlockDirective("panel", "", func(expression string, body string) (string, error) {
			return `<section class="panel" data-kind="` + expression + `">` + body + `</section>`, nil
		}),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"title": "Blade HTTP Demo",
			"brand": "sparq/blade",
			"path":  r.URL.Path,
			"now":   time.Now().Format(time.RFC3339),
			"features": []map[string]any{
				{"name": "extends/include", "ready": true},
				{"name": "component/slot", "ready": true},
				{"name": "each/custom directives", "ready": true},
			},
		}

		renderHTML(w, renderer, "home", data)
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"title": "Users",
			"brand": "sparq/blade",
			"path":  r.URL.Path,
			"users": []map[string]any{
				{"name": "Alice", "active": true},
				{"name": "Bob", "active": false},
				{"name": "Carol", "active": true},
			},
		}

		renderHTML(w, renderer, "users", data)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := "127.0.0.1:8080"
	log.Printf("blade http demo listening on http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func renderHTML(w http.ResponseWriter, renderer *blade.Engine, view string, data any) {
	output, err := renderer.RenderFile(view, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(output))
}

func viewsRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "views")
}
