package main

import (
	"fmt"
	"path/filepath"
	"runtime"

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

	output, err := renderer.RenderFile("page", map[string]any{
		"users": []map[string]any{
			{"name": "Alice", "active": true},
			{"name": "Bob", "active": false},
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)
}

func viewsRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "views")
}
