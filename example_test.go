package blade_test

import (
	"fmt"
	"os"
	"path/filepath"

	blade "github.com/zane-xin/sparq-blade"
)

func ExampleRenderString() {
	output, err := blade.RenderString(`<h1>{{ title }}</h1>`, map[string]any{
		"title": "Blade",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(output)
	// Output:
	// <h1>Blade</h1>
}

func ExampleNew_withDirective() {
	renderer := blade.New(
		blade.WithDirective("badge", func(expression string) (string, error) {
			return fmt.Sprintf(`<strong>{{ bladeEcho . %q }}</strong>`, expression), nil
		}),
	)

	output, err := renderer.RenderString(`@sq:badge(name)`, map[string]any{
		"name": "Blade",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(output)
	// Output:
	// <strong>Blade</strong>
}

func ExampleNew_withBlockDirective() {
	renderer := blade.New(
		blade.WithBlockDirective("panel", "", func(_ string, body string) (string, error) {
			return `<section class="panel">` + body + `</section>`, nil
		}),
	)

	output, err := renderer.RenderString(`@sq:panel()Hello {{ name }}@end`, map[string]any{
		"name": "Blade",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(output)
	// Output:
	// <section class="panel">Hello Blade</section>
}

func ExampleNew_withFS() {
	root, err := os.MkdirTemp("", "blade-example-*")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer os.RemoveAll(root)

	mustWriteExample(filepath.Join(root, "layout.tpl"), `<html><body>@yield('content')</body></html>`)
	mustWriteExample(filepath.Join(root, "page.tpl"), `@extends('layout')@section('content')<p>{{ name }}-{{ count }}</p>@endsection`)

	renderer := blade.New(
		blade.WithFS(blade.OSFS{Root: root}),
	)

	output, err := renderer.RenderFile("page", map[string]any{
		"name":  "sq",
		"count": 2,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(output)
	// Output:
	// <html><body><p>sq-2</p></body></html>
}

func mustWriteExample(path string, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		panic(err)
	}
}
