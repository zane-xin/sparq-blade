package blade_test

import (
	"os"
	"path/filepath"
	"testing"

	blade "github.com/zane-xin/sparq-blade"
)

func BenchmarkBlade(b *testing.B) {
	parseSource := `@extends('layout')@section('content')<ul>@foreach(items as item)@component('card')@slot('title'){{ item.title }}@endslot@if(item.visible)<b>{{ item.name }}</b>@endif@endcomponent@endforeach</ul>@endsection`
	renderStringSource := `<ul>@foreach(items as item)<li>@if(item.visible){{ item.title }}-{{ item.name }}@endif</li>@endforeach</ul>`

	data := map[string]any{
		"items": []map[string]any{
			{"title": "A", "name": "Alpha", "visible": true},
			{"title": "B", "name": "Beta", "visible": true},
			{"title": "C", "name": "Gamma", "visible": false},
		},
	}

	b.Run("Lex", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := blade.Lex(parseSource); err != nil {
				b.Fatalf("Lex() error = %v", err)
			}
		}
	})

	b.Run("Parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := blade.Parse(parseSource); err != nil {
				b.Fatalf("Parse() error = %v", err)
			}
		}
	})

	b.Run("RenderString", func(b *testing.B) {
		b.ReportAllocs()
		renderer := blade.New()
		for i := 0; i < b.N; i++ {
			if _, err := renderer.RenderString(renderStringSource, data); err != nil {
				b.Fatalf("RenderString() error = %v", err)
			}
		}
	})

	b.Run("RenderFile", func(b *testing.B) {
		b.ReportAllocs()
		root := b.TempDir()
		writeBenchmarkFile(b, filepath.Join(root, "layout.tpl"), `<html><body>@yield('content')</body></html>`)
		writeBenchmarkFile(b, filepath.Join(root, "card.tpl"), `<article><header>{{ title }}</header><section>{{ slot }}</section></article>`)
		writeBenchmarkFile(b, filepath.Join(root, "page.tpl"), parseSource)

		renderer := blade.New(
			blade.WithFS(blade.OSFS{Root: root}),
		)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := renderer.RenderFile("page", data); err != nil {
				b.Fatalf("RenderFile() error = %v", err)
			}
		}
	})
}

func writeBenchmarkFile(b *testing.B, path string, content string) {
	b.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
