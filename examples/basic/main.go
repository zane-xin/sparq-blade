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
	)

	output, err := renderer.RenderFile("home", map[string]any{
		"title": "Blade Basic Example",
		"items": []map[string]any{
			{"name": "Parser"},
			{"name": "Engine"},
			{"name": "Runtime"},
		},
		"brand": "sparq/blade",
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
