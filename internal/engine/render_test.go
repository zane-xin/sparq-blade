package engine

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/zane-xin/sparq-blade/internal/compiler"
)

func TestEngineRenderString(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`<h1>{{ title }}</h1>@if(visible)<p>{!! html !!}</p>@else<p>hidden</p>@endif`,
		map[string]any{
			"title":   "<Blade>",
			"visible": true,
			"html":    "<strong>safe</strong>",
		},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "<h1>&lt;Blade&gt;</h1><p><strong>safe</strong></p>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderStringWithCustomSyntax(t *testing.T) {
	engine := New(WithSyntax(compiler.Syntax{
		EchoOpen:     "[[",
		EchoClose:    "]]",
		RawOpen:      "[[!",
		RawClose:     "!]]",
		CommentOpen:  "[[--",
		CommentClose: "--]]",
	}))

	output, err := engine.RenderString(
		`<h1>[[ title ]]</h1>@if(visible)<p>[[! html !]]</p>[[-- hidden --]]@else<p>hidden</p>@endif`,
		map[string]any{
			"title":   "<Blade>",
			"visible": true,
			"html":    "<strong>safe</strong>",
		},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "<h1>&lt;Blade&gt;</h1><p><strong>safe</strong></p>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderForeachIncludeAndExtends(t *testing.T) {
	root := t.TempDir()

	writeTemplate(t, root, "layout.tpl", `<html><body>@yield('content')</body></html>`)
	writeTemplate(t, root, "footer.tpl", `<footer>{{ brand }}</footer>`)
	writeTemplate(t, root, "home.tpl", `@extends('layout')@section('content')<ul>@foreach(items as item)<li>{{ item.name }}</li>@endforeach</ul>@include('footer')@endsection`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("home", map[string]any{
		"items": []map[string]any{
			{"name": "A"},
			{"name": "B"},
		},
		"brand": "Sparq",
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<html><body><ul><li>A</li><li>B</li></ul><footer>Sparq</footer></body></html>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderIncludeWithData(t *testing.T) {
	root := t.TempDir()

	writeTemplate(t, root, "card.tpl", `<article><h2>{{ name }}</h2><p>{{ brand }}</p><span>{{ role }}</span></article>`)
	writeTemplate(t, root, "page.tpl", `@include('card', user)`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("page", map[string]any{
		"brand": "Sparq",
		"user": map[string]any{
			"name": "Alice",
			"role": "admin",
		},
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<article><h2>Alice</h2><p>Sparq</p><span>admin</span></article>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderIncludeWithDataKeepsLoopScope(t *testing.T) {
	root := t.TempDir()

	writeTemplate(t, root, "row.tpl", `<li>{{ item.name }}-{{ name }}</li>`)
	writeTemplate(t, root, "page.tpl", `<ul>@foreach(users as item)@include('row', profile)@endforeach</ul>`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("page", map[string]any{
		"profile": map[string]any{
			"name": "Team",
		},
		"users": []map[string]any{
			{"name": "A"},
			{"name": "B"},
		},
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<ul><li>A-Team</li><li>B-Team</li></ul>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderFileCacheInvalidation(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "page.tpl", `<p>{{ value }}</p>`)

	engine := New(WithFS(OSFS{Root: root}))

	first, err := engine.RenderFile("page", map[string]any{"value": "old"})
	if err != nil {
		t.Fatalf("RenderFile() first error = %v", err)
	}
	if first != "<p>old</p>" {
		t.Fatalf("unexpected first output: %q", first)
	}

	path := filepath.Join(root, "page.tpl")
	if err := os.WriteFile(path, []byte(`<div>{{ value }}</div>`), 0o644); err != nil {
		t.Fatalf("rewrite template error = %v", err)
	}

	second, err := engine.RenderFile("page", map[string]any{"value": "new"})
	if err != nil {
		t.Fatalf("RenderFile() second error = %v", err)
	}
	if second != "<div>new</div>" {
		t.Fatalf("unexpected second output: %q", second)
	}
}

func TestEngineRenderFileCacheInvalidatesWhenParentChanges(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "layout.tpl", `<html><body>@yield('content')</body></html>`)
	writeTemplate(t, root, "page.tpl", `@extends('layout')@section('content')<p>{{ value }}</p>@endsection`)

	engine := New(WithFS(OSFS{Root: root}))

	first, err := engine.RenderFile("page", map[string]any{"value": "old"})
	if err != nil {
		t.Fatalf("RenderFile() first error = %v", err)
	}
	if first != "<html><body><p>old</p></body></html>" {
		t.Fatalf("unexpected first output: %q", first)
	}

	path := filepath.Join(root, "layout.tpl")
	if err := os.WriteFile(path, []byte(`<main>@yield('content')</main>`), 0o644); err != nil {
		t.Fatalf("rewrite layout error = %v", err)
	}

	second, err := engine.RenderFile("page", map[string]any{"value": "new"})
	if err != nil {
		t.Fatalf("RenderFile() second error = %v", err)
	}
	if second != "<main><p>new</p></main>" {
		t.Fatalf("unexpected second output: %q", second)
	}
}

func TestOSFSLoadSupportsOnlyTplAndHTML(t *testing.T) {
	root := t.TempDir()

	writeTemplate(t, root, "page.tpl", `<p>tpl</p>`)
	writeTemplate(t, root, "page.html", `<p>html</p>`)
	writeTemplate(t, root, "legacy.blade.php", `<p>legacy</p>`)
	writeTemplate(t, root, "legacy.blade", `<p>legacy</p>`)
	writeTemplate(t, root, "legacy.tmpl", `<p>legacy</p>`)

	fs := OSFS{Root: root}

	for _, name := range []string{"page", "page.tpl", "page.html"} {
		if _, err := fs.Load(name); err != nil {
			t.Fatalf("Load(%q) error = %v", name, err)
		}
	}

	for _, name := range []string{"legacy", "legacy.blade.php", "legacy.blade", "legacy.tmpl"} {
		if _, err := fs.Load(name); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Load(%q) expected os.ErrNotExist, got %v", name, err)
		}
	}
}

func TestOSFSLoadRejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	cmsRoot := filepath.Join(root, "cms")

	writeTemplate(t, root, "pages/home.tpl", `<p>home</p>`)
	writeTemplate(t, cmsRoot, "pages/home.tpl", `<p>cms</p>`)
	writeTemplate(t, outside, "secret.tpl", `<p>secret</p>`)

	fs := OSFS{
		Root: root,
		Namespaces: map[string]string{
			"cms": cmsRoot,
		},
	}

	for _, name := range []string{
		"../secret",
		"../../secret.tpl",
		"/pages/home",
		string(filepath.Separator) + "pages" + string(filepath.Separator) + "home",
		filepath.Join(outside, "secret.tpl"),
		"cms:../pages/home",
		"cms:../../secret.tpl",
		"cms:/pages/home",
		"cms:" + string(filepath.Separator) + "pages" + string(filepath.Separator) + "home",
		"cms:" + filepath.Join(outside, "secret.tpl"),
	} {
		if _, err := fs.Load(name); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Load(%q) expected os.ErrNotExist, got %v", name, err)
		}
	}

	for _, name := range []string{"pages.home", "cms:pages.home"} {
		if _, err := fs.Load(name); err != nil {
			t.Fatalf("Load(%q) error = %v", name, err)
		}
	}
}

func TestOSFSLoadRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	writeTemplate(t, root, "real.tpl", `<p>real</p>`)
	writeTemplate(t, outside, "secret.tpl", `<p>secret</p>`)

	insideLink := filepath.Join(root, "inside.tpl")
	if err := os.Symlink(filepath.Join(root, "real.tpl"), insideLink); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	outsideLink := filepath.Join(root, "outside.tpl")
	if err := os.Symlink(filepath.Join(outside, "secret.tpl"), outsideLink); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	fs := OSFS{Root: root}

	if _, err := fs.Load("inside"); err != nil {
		t.Fatalf("Load(%q) error = %v", "inside", err)
	}

	for _, name := range []string{"outside", "outside.tpl"} {
		if _, err := fs.Load(name); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Load(%q) expected os.ErrNotExist, got %v", name, err)
		}
	}
}

func TestEngineRenderEscapedDirectiveLiteral(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(`@@sq:title @@if(ok)`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `@sq:title @if(ok)`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestLayeredFSLoadPrefersFirstLayer(t *testing.T) {
	themeRoot := t.TempDir()
	fallbackRoot := t.TempDir()
	writeTemplate(t, themeRoot, "home.tpl", "theme")
	writeTemplate(t, fallbackRoot, "home.tpl", "fallback")

	fs := LayeredFS{
		Layers: []TemplateFS{
			OSFS{Root: themeRoot},
			OSFS{Root: fallbackRoot},
		},
	}

	source, err := fs.Load("home")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if string(source.Content) != "theme" {
		t.Fatalf("unexpected content: %q", source.Content)
	}
	if !strings.HasPrefix(source.Name, "layer0:") {
		t.Fatalf("expected layer0 source name, got %q", source.Name)
	}
}

func TestLayeredFSLoadFallsBackToNextLayer(t *testing.T) {
	themeRoot := t.TempDir()
	fallbackRoot := t.TempDir()
	writeTemplate(t, fallbackRoot, "home.tpl", "fallback")

	fs := LayeredFS{
		Layers: []TemplateFS{
			OSFS{Root: themeRoot},
			OSFS{Root: fallbackRoot},
		},
	}

	source, err := fs.Load("home")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if string(source.Content) != "fallback" {
		t.Fatalf("unexpected content: %q", source.Content)
	}
	if !strings.HasPrefix(source.Name, "layer1:") {
		t.Fatalf("expected layer1 source name, got %q", source.Name)
	}
}

func TestLayeredFSLoadSupportsNamespacesByLayer(t *testing.T) {
	themeRoot := t.TempDir()
	moduleRoot := t.TempDir()
	writeTemplate(t, themeRoot, "cms/partials/pager.tpl", "theme pager")
	writeTemplate(t, moduleRoot, "partials/pager.tpl", "module pager")

	fs := LayeredFS{
		Layers: []TemplateFS{
			OSFS{
				Root: themeRoot,
				Namespaces: map[string]string{
					"cms": filepath.Join(themeRoot, "cms"),
				},
			},
			OSFS{
				Root: moduleRoot,
				Namespaces: map[string]string{
					"cms": moduleRoot,
				},
			},
		},
	}

	source, err := fs.Load("cms:partials.pager")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if string(source.Content) != "theme pager" {
		t.Fatalf("unexpected content: %q", source.Content)
	}
	if !strings.HasPrefix(source.Name, "layer0:cms:") {
		t.Fatalf("expected layer0 cms source name, got %q", source.Name)
	}
}

func TestLayeredFSLoadReturnsFirstNonNotExistError(t *testing.T) {
	expectedErr := errors.New("boom")
	fs := LayeredFS{
		Layers: []TemplateFS{
			errorFS{err: expectedErr},
			OSFS{Root: t.TempDir()},
		},
	}

	_, err := fs.Load("home")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestLayeredFSLoadRejectsPathTraversal(t *testing.T) {
	themeRoot := t.TempDir()
	fallbackRoot := t.TempDir()
	writeTemplate(t, fallbackRoot, "home.tpl", "fallback")

	fs := LayeredFS{
		Layers: []TemplateFS{
			OSFS{Root: themeRoot},
			OSFS{Root: fallbackRoot},
		},
	}

	for _, name := range []string{"../home", "partials/../../home", "/home"} {
		if _, err := fs.Load(name); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Load(%q) expected os.ErrNotExist, got %v", name, err)
		}
	}
}

func TestEngineRenderLayeredFSUsesLayeredCacheName(t *testing.T) {
	themeRoot := t.TempDir()
	fallbackRoot := t.TempDir()
	writeTemplate(t, fallbackRoot, "page.tpl", `<p>fallback {{ value }}</p>`)

	engine := New(WithFS(LayeredFS{
		Layers: []TemplateFS{
			OSFS{Root: themeRoot},
			OSFS{Root: fallbackRoot},
		},
	}))

	first, err := engine.RenderFile("page", map[string]any{"value": "A"})
	if err != nil {
		t.Fatalf("RenderFile() first error = %v", err)
	}
	if first != "<p>fallback A</p>" {
		t.Fatalf("unexpected first output: %q", first)
	}

	writeTemplate(t, themeRoot, "page.tpl", `<p>theme {{ value }}</p>`)
	second, err := engine.RenderFile("page", map[string]any{"value": "B"})
	if err != nil {
		t.Fatalf("RenderFile() second error = %v", err)
	}
	if second != "<p>theme B</p>" {
		t.Fatalf("unexpected second output: %q", second)
	}
}

func TestEngineRenderTagContextMetaPriority(t *testing.T) {
	type metaContext struct {
		System     map[string]string
		Route      map[string]string
		Controller map[string]string
		Page       map[string]string
	}

	resolve := func(ctx metaContext, key string) string {
		for _, source := range []map[string]string{ctx.Page, ctx.Controller, ctx.Route, ctx.System} {
			if value := source[key]; value != "" {
				return value
			}
		}
		return ""
	}

	engine := New(
		WithTag("title", func(call TagCall) (template.HTML, error) {
			ctx := call.Context().(metaContext)
			return template.HTML(template.HTMLEscapeString(resolve(ctx, "title"))), nil
		}),
		WithTag("seo", func(call TagCall) (template.HTML, error) {
			ctx := call.Context().(metaContext)
			return template.HTML(fmt.Sprintf(
				`<meta name="keywords" content="%s"><meta name="description" content="%s">`,
				template.HTMLEscapeString(resolve(ctx, "keywords")),
				template.HTMLEscapeString(resolve(ctx, "description")),
			)), nil
		}),
	)

	output, err := engine.RenderString(`<title>@sq:title</title>@sq:seo`, metaContext{
		System: map[string]string{
			"title":       "系统标题",
			"keywords":    "系统关键词",
			"description": "系统描述",
		},
		Route: map[string]string{
			"title":       "栏目标题",
			"description": "栏目描述",
		},
		Controller: map[string]string{
			"title":    "控制器标题",
			"keywords": "控制器关键词",
		},
		Page: map[string]string{
			"title": "页面标题",
		},
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<title>页面标题</title><meta name="keywords" content="控制器关键词"><meta name="description" content="栏目描述">`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderDataTagInfersCategoryFromContext(t *testing.T) {
	type categoryContext struct {
		CategoryID int
	}

	engine := New(
		WithDataDirective("arc", func(call DataTagCall) (DataTagResult, error) {
			ctx := call.Context().(categoryContext)
			cid, err := call.GetIntDefault("cid", ctx.CategoryID)
			if err != nil {
				return DataTagResult{}, err
			}

			return Items(Item(Fields().Set("title", fmt.Sprintf("cid-%d", cid)))), nil
		}),
	)

	output, err := engine.RenderString(
		`<ul>@sq:arc()<li>{{ field.title }}</li>@end @sq:arc(cid=99)<li>{{ field.title }}</li>@end</ul>`,
		categoryContext{CategoryID: 7},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<ul><li>cid-7</li><li>cid-99</li></ul>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderFileConcurrent(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "layout.tpl", `<html><body>@yield('content')</body></html>`)
	writeTemplate(t, root, "page.tpl", `@extends('layout')@section('content')<p>{{ value }}</p>@endsection`)

	engine := New(WithFS(OSFS{Root: root}))

	var wg sync.WaitGroup
	errs := make(chan error, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			output, err := engine.RenderFile("page", map[string]any{
				"value": fmt.Sprintf("item-%d", index),
			})
			if err != nil {
				errs <- err
				return
			}

			expected := fmt.Sprintf("<html><body><p>item-%d</p></body></html>", index)
			if output != expected {
				errs <- fmt.Errorf("unexpected output: got %q want %q", output, expected)
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestEngineRenderEach(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "row.tpl", `<li>{{ user.name }}</li>`)
	writeTemplate(t, root, "empty.tpl", `<p>empty</p>`)
	writeTemplate(t, root, "list.tpl", `<ul>@each('row', users, 'user', 'empty')</ul>`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("list", map[string]any{
		"users": []map[string]any{
			{"name": "A"},
			{"name": "B"},
		},
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<ul><li>A</li><li>B</li></ul>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}

	emptyOutput, err := engine.RenderFile("list", map[string]any{
		"users": []map[string]any{},
	})
	if err != nil {
		t.Fatalf("RenderFile() empty error = %v", err)
	}

	expectedEmpty := "<ul><p>empty</p></ul>"
	if emptyOutput != expectedEmpty {
		t.Fatalf("unexpected empty output: got %q want %q", emptyOutput, expectedEmpty)
	}
}

func TestEngineRenderForeachLoopState(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@foreach(items as item)<span>{{ loop.index }}:{{ loop.iteration }}:{{ loop.remaining }}:{{ loop.count }}:{{ loop.first }}:{{ loop.last }}:{{ loop.odd }}:{{ loop.even }}:{{ item }}</span>@endforeach`,
		map[string]any{"items": []string{"A", "B"}},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<span>0:1:1:2:true:false:true:false:A</span><span>1:2:0:2:false:true:false:true:B</span>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderComponentWithSlots(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "card.tpl", `<article><header>{{ title }}</header><section>{{ slot }}</section><footer>{{ footer }}</footer></article>`)
	writeTemplate(t, root, "page.tpl", `@component('card')Body {{ name }}@slot('title')Hello {{ name }}@endslot@slot('footer')Footer@endslot@endcomponent`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("page", map[string]any{
		"name": "Blade",
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<article><header>Hello Blade</header><section>Body Blade</section><footer>Footer</footer></article>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderComponentWithData(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "card.tpl", `<article><h2>{{ title }}</h2><section>{{ slot }}</section><footer>{{ brand }}</footer></article>`)
	writeTemplate(t, root, "page.tpl", `@component('card', card)Body {{ name }}@endcomponent`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("page", map[string]any{
		"brand": "Sparq",
		"name":  "Outer",
		"card": map[string]any{
			"title": "Card Title",
		},
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<article><h2>Card Title</h2><section>Body Outer</section><footer>Sparq</footer></article>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderComponentWithDataKeepsLoopScope(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "chip.tpl", `<span>{{ item.name }}:{{ kind }}:{{ slot }}</span>`)
	writeTemplate(t, root, "page.tpl", `@foreach(users as item)@component('chip', meta){{ item.name }}@endcomponent@endforeach`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("page", map[string]any{
		"meta": map[string]any{
			"kind": "user",
		},
		"users": []map[string]any{
			{"name": "A"},
			{"name": "B"},
		},
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<span>A:user:A</span><span>B:user:B</span>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderComponentWithBladeContent(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "card.tpl", `<article><header>{{ title }}</header><section>{{ slot }}</section><footer>{{ footer }}</footer></article>`)
	writeTemplate(t, root, "page.tpl", `@component('card')@if(visible)<b>{{ name }}</b>@endif@slot('title')Hello {{ name }}@endslot@slot('footer')@if(visible)Footer {{ name }}@endif@endslot@endcomponent`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("page", map[string]any{
		"visible": true,
		"name":    "Blade",
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<article><header>Hello Blade</header><section><b>Blade</b></section><footer>Footer Blade</footer></article>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderCustomDirective(t *testing.T) {
	engine := New(
		WithDirective("badge", func(expression string) (string, error) {
			return fmt.Sprintf(`<strong>{{ bladeEcho . %q }}</strong>`, expression), nil
		}),
	)

	output, err := engine.RenderString(`@sq:badge(name)`, map[string]any{
		"name": "Blade",
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "<strong>Blade</strong>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderBlockDirective(t *testing.T) {
	engine := New(
		WithBlockDirective("panel", "", func(_ string, body string) (string, error) {
			return `<section class="panel">` + body + `</section>`, nil
		}),
	)

	output, err := engine.RenderString(`@sq:panel()Hello {{ name }}@end`, map[string]any{
		"name": "Blade",
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<section class="panel">Hello Blade</section>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderBlockDirectiveWithConditionalBody(t *testing.T) {
	engine := New(
		WithBlockDirective("panel", "", func(expression string, body string) (string, error) {
			return `<section class="panel" data-name="` + expression + `">` + body + `</section>`, nil
		}),
	)

	output, err := engine.RenderString(
		`@sq:panel(primary)@if(visible)<b>{{ name }}</b>@endif@end`,
		map[string]any{
			"visible": true,
			"name":    "Blade",
		},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<section class="panel" data-name="primary"><b>Blade</b></section>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderTagDirective(t *testing.T) {
	engine := New(
		WithTag("snippet", func(call TagCall) (template.HTML, error) {
			return template.HTML(fmt.Sprintf(`<div data-target="%s" data-limit="%v" data-pid="%v"></div>`, call.Target, call.Params["limit"], call.Params["pid"])), nil
		}),
	)

	output, err := engine.RenderString(`@sq:snippet('cms:arc', limit=1, pid=pid)`, map[string]any{
		"pid": 9,
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<div data-target="cms:arc" data-limit="1" data-pid="9"></div>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderTagBlockDirective(t *testing.T) {
	engine := New(
		WithTag("snippet", func(call TagCall) (template.HTML, error) {
			return template.HTML(fmt.Sprintf(`<section data-target="%s" data-kind="%v">%s</section>`, call.Target, call.Params["kind"], call.Body)), nil
		}),
	)

	output, err := engine.RenderString(`@sq:snippet('cms:list', kind='news')<b>{{ title }}</b>@sq:endsnippet`, map[string]any{
		"title": "Blade",
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<section data-target="cms:list" data-kind="news"><b>Blade</b></section>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderTagCallParamHelpers(t *testing.T) {
	engine := New(
		WithTag("pager", func(call TagCall) (template.HTML, error) {
			if !call.Has("size") || call.Has("missing") {
				return "", fmt.Errorf("unexpected param presence")
			}

			size, err := call.GetIntDefault("size", 10)
			if err != nil {
				return "", err
			}
			className, err := call.GetStringDefault("class", "pager")
			if err != nil {
				return "", err
			}
			enabled, err := call.GetBoolDefault("enabled", false)
			if err != nil {
				return "", err
			}

			return template.HTML(fmt.Sprintf(`<nav class="%s" data-size="%d" data-enabled="%t"></nav>`, template.HTMLEscapeString(className), size, enabled)), nil
		}),
	)

	output, err := engine.RenderString(`@sq:pager(size=20, class="page-links", enabled=true)`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<nav class="page-links" data-size="20" data-enabled="true"></nav>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderBusinessTagBlock(t *testing.T) {
	engine := New(
		WithTag("badge", func(call TagCall) (template.HTML, error) {
			return template.HTML(fmt.Sprintf(`<span>%s</span>`, call.Body)), nil
		}),
	)

	output, err := engine.RenderString(`@sq:badge()<b>{{ title }}</b>@sq:endbadge`, map[string]any{
		"title": "New",
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<span><b>New</b></span>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineWithTagRejectsBuiltinDirectiveName(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic, got nil")
		}
		if !strings.Contains(fmt.Sprint(recovered), "业务标签指令名与内置指令冲突: @sq:list") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	New(WithTag("list", func(call TagCall) (template.HTML, error) {
		return "", nil
	}))
}

func TestEngineWithTagRejectsBrandedBuiltinDirectiveName(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic, got nil")
		}
		if !strings.Contains(fmt.Sprint(recovered), "业务指令注册只接受局部名称: sq:children") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	New(WithTag("sq:children", func(call TagCall) (template.HTML, error) {
		return "", nil
	}))
}

func TestEngineWithDataDirectiveRejectsBuiltinDirectiveName(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic, got nil")
		}
		if !strings.Contains(fmt.Sprint(recovered), "数据标签指令名与内置指令冲突: @sq:list") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	New(WithDataDirective("list", func(call DataTagCall) (DataTagResult, error) {
		return EmptyResult(), nil
	}))
}

func TestEngineRejectsTagAndDataDirectiveNameConflict(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic, got nil")
		}
		if !strings.Contains(fmt.Sprint(recovered), "数据标签指令名已被注册: @sq:pager") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	New(
		WithTag("pager", func(call TagCall) (template.HTML, error) {
			return "", nil
		}),
		WithDataDirective("pager", func(call DataTagCall) (DataTagResult, error) {
			return EmptyResult(), nil
		}),
	)
}

func TestEngineRejectsCustomDirectiveNameConflict(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic, got nil")
		}
		if !strings.Contains(fmt.Sprint(recovered), "业务标签指令名已被注册: @sq:pager") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	New(
		WithDirective("pager", func(expression string) (string, error) {
			return "", nil
		}),
		WithTag("pager", func(call TagCall) (template.HTML, error) {
			return "", nil
		}),
	)
}

func TestEngineRejectsBlockDirectiveEndNameConflict(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic, got nil")
		}
		if !strings.Contains(fmt.Sprint(recovered), "业务指令名已被注册: @sq:close") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	New(
		WithBlockDirective("panel", "close", func(expression string, body string) (string, error) {
			return "", nil
		}),
		WithDirective("close", func(expression string) (string, error) {
			return "", nil
		}),
	)
}

func TestEngineRejectsInvalidDirectiveName(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic, got nil")
		}
		if !strings.Contains(fmt.Sprint(recovered), "业务标签指令名非法") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	New(WithTag("", func(call TagCall) (template.HTML, error) {
		return "", nil
	}))
}

func TestEngineRenderTagInfersPagerFromContext(t *testing.T) {
	type pagerContext struct {
		Page        map[string]any
		CustomPager map[string]any
	}

	engine := New(
		WithTag("pager", func(call TagCall) (template.HTML, error) {
			pager, ok := call.Get("for")
			if !ok {
				ctx := call.Context().(pagerContext)
				pager = ctx.Page["pager"]
			}

			info := pager.(map[string]any)
			className, err := call.GetStringDefault("class", "pager")
			if err != nil {
				return "", err
			}

			return template.HTML(fmt.Sprintf(
				`<nav class="%s" data-page="%v" data-total="%v"></nav>`,
				template.HTMLEscapeString(className),
				info["page"],
				info["total"],
			)), nil
		}),
	)

	output, err := engine.RenderString(`@sq:pager() @sq:pager(for=customPager, class="mini")`, pagerContext{
		Page: map[string]any{
			"pager": map[string]any{"page": 1, "total": 30},
		},
		CustomPager: map[string]any{"page": 2, "total": 80},
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<nav class="pager" data-page="1" data-total="30"></nav> <nav class="mini" data-page="2" data-total="80"></nav>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderTagCallRenderTemplate(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "page.tpl", `@sq:pager(template="partials.pager", class="compact")`)
	writeTemplate(t, root, "partials/pager.tpl", `<nav class="{{ class }}" data-site="{{ site }}">{{ pager.page }}/{{ pager.total }}</nav>`)

	type pageContext struct {
		Site  string
		Pager map[string]any
	}

	engine := New(
		WithFS(OSFS{Root: root}),
		WithTag("pager", func(call TagCall) (template.HTML, error) {
			view, err := call.GetStringDefault("template", "")
			if err != nil {
				return "", err
			}
			className, err := call.GetStringDefault("class", "pager")
			if err != nil {
				return "", err
			}
			if view == "" {
				return "", fmt.Errorf("missing template")
			}

			ctx := call.Context().(pageContext)
			return call.Render(view, map[string]any{
				"pager": ctx.Pager,
				"class": className,
			})
		}),
	)

	output, err := engine.RenderFile("page", pageContext{
		Site:  "docs",
		Pager: map[string]any{"page": 3, "total": 9},
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := `<nav class="compact" data-site="docs">3/9</nav>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderTagDirectiveInvalidTarget(t *testing.T) {
	engine := New(
		WithTag("snippet", func(call TagCall) (template.HTML, error) {
			return template.HTML(call.Target), nil
		}),
	)

	_, err := engine.RenderString(`@sq:snippet(target, limit=1)`, map[string]any{
		"target": "cms:arc",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "tag 目标必须是字符串字面量") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineRenderDataTagBlock(t *testing.T) {
	engine := New(
		WithDataDirective("arc", func(call DataTagCall) (DataTagResult, error) {
			pid, err := call.GetInt("pid")
			if err != nil {
				return DataTagResult{}, err
			}
			limit, err := call.GetIntDefault("limit", 10)
			if err != nil {
				return DataTagResult{}, err
			}
			if !call.Has("as") {
				return DataTagResult{}, fmt.Errorf("expected as param")
			}
			if pid != 10 || limit != 2 {
				return DataTagResult{}, fmt.Errorf("unexpected params: pid=%d limit=%d", pid, limit)
			}

			return DataTagResult{
				Items: []any{
					map[string]any{"id": 1, "title": "A", "url": "/a"},
					map[string]any{"id": 2, "title": "B", "url": "/b"},
				},
			}, nil
		}),
	)

	output, err := engine.RenderString(
		`<ul>@sq:arc(pid=10, limit='2', as=field)<li>{{ loop.index }}-{{ loop.remainini }}-@if(loop.odd)odd@endif@if(loop.even)even@endif-<a href="{{ field.url }}">{{ field.title }}</a>@if(field.id == 1)first@endif</li>@empty<li>empty</li>@end</ul>`,
		nil,
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<ul><li>0-1-odd-<a href="/a">A</a>first</li><li>1-0-even-<a href="/b">B</a></li></ul>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderDataTagOut(t *testing.T) {
	engine := New(
		WithDataDirective("arc", func(call DataTagCall) (DataTagResult, error) {
			out, ok, err := call.Out()
			if err != nil {
				return DataTagResult{}, err
			}
			if !ok || out != "arc" {
				return DataTagResult{}, fmt.Errorf("unexpected out: %q %v", out, ok)
			}
			active, err := call.GetBoolDefault("active", false)
			if err != nil {
				return DataTagResult{}, err
			}
			if !active {
				return DataTagResult{}, fmt.Errorf("expected active=true")
			}

			return DataTagResult{
				Items: []any{
					map[string]any{"title": "A"},
					map[string]any{"title": "B"},
				},
			}, nil
		}),
	)

	output, err := engine.RenderString(
		`@sq:arc(pid=10, active='true', out='arc')<ul>@foreach(arc as item)<li>{{ item.title }}</li>@endforeach</ul>`,
		nil,
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<ul><li>A</li><li>B</li></ul>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderDataTagOutShorthand(t *testing.T) {
	engine := New(
		WithDataDirective("arc", func(call DataTagCall) (DataTagResult, error) {
			out, ok, err := call.Out()
			if err != nil {
				return DataTagResult{}, err
			}
			if !ok || out != "field" {
				return DataTagResult{}, fmt.Errorf("unexpected out: %q %v", out, ok)
			}

			return DataTagResult{
				Items: []any{
					map[string]any{"title": "A"},
					map[string]any{"title": "B"},
				},
			}, nil
		}),
	)

	output, err := engine.RenderString(
		`@sq:arc(pid=10, out)<ul>@foreach(field as item)<li>{{ item.title }}</li>@endforeach</ul>`,
		nil,
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<ul><li>A</li><li>B</li></ul>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderDataTagOutValue(t *testing.T) {
	engine := New(
		WithDataDirective("site", func(call DataTagCall) (DataTagResult, error) {
			out, ok, err := call.Out()
			if err != nil {
				return DataTagResult{}, err
			}
			if !ok || out != "field" {
				return DataTagResult{}, fmt.Errorf("unexpected out: %q %v", out, ok)
			}

			return DataTagResult{
				Value: map[string]any{
					"title": "SQ CMS",
					"url":   "https://example.test",
				},
			}, nil
		}),
	)

	output, err := engine.RenderString(
		`@sq:site(out)<h1>{{ field.title }}</h1><a href="{{ field.url }}">link</a>`,
		nil,
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<h1>SQ CMS</h1><a href="https://example.test">link</a>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderDataTagSingleBeforeBlockInSameSection(t *testing.T) {
	engine := New(
		WithDataDirective("arc", func(call DataTagCall) (DataTagResult, error) {
			cid, err := call.GetIntDefault("cid", 1)
			if err != nil {
				return DataTagResult{}, err
			}

			items := []any{
				map[string]any{"title": fmt.Sprintf("item-%d-a", cid)},
				map[string]any{"title": fmt.Sprintf("item-%d-b", cid)},
			}

			if _, ok, err := call.Out(); err != nil {
				return DataTagResult{}, err
			} else if ok {
				return ValueOf(items), nil
			}

			return Items(items...), nil
		}),
	)

	output, err := engine.RenderString(
		`@section('content')@sq:arc(cid=7, out=articles)@foreach(articles as item)<b>{{ item.title }}</b>@endforeach @sq:arc(cid=9)<i>{{ field.title }}</i>@end@endsection`,
		nil,
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<b>item-7-a</b><b>item-7-b</b><i>item-9-a</i><i>item-9-b</i>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderDataTagRejectsBrandedEndDirective(t *testing.T) {
	engine := New(
		WithDataDirective("arc", func(call DataTagCall) (DataTagResult, error) {
			return Items(map[string]any{"title": "A"}), nil
		}),
	)

	_, err := engine.RenderString(`@sq:arc(as=article){{ article.title }}@sq:endarc`, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "当前版本尚未支持该指令: @sq:endarc") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineRenderDataTagBlockRejectsValue(t *testing.T) {
	engine := New(
		WithDataDirective("site", func(call DataTagCall) (DataTagResult, error) {
			return DataTagResult{
				Value: map[string]any{"title": "SQ CMS"},
			}, nil
		}),
	)

	_, err := engine.RenderString(`@sq:site(as=field){{ field.title }}@end`, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "@sq 块标签只支持列表结果") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineRenderDataTagOutFalseInvalid(t *testing.T) {
	engine := New(
		WithDataDirective("arc", func(call DataTagCall) (DataTagResult, error) {
			return DataTagResult{}, nil
		}),
	)

	_, err := engine.RenderString(`@sq:arc(out=false)`, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "参数 out=false 无效") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineRenderDataTagOutRejectsMixedValueAndItems(t *testing.T) {
	engine := New(
		WithDataDirective("site", func(call DataTagCall) (DataTagResult, error) {
			return DataTagResult{
				Value: map[string]any{"title": "SQ CMS"},
				Items: []any{map[string]any{"title": "A"}},
			}, nil
		}),
	)

	_, err := engine.RenderString(`@sq:site(out='site')`, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "不能同时设置 Value 和 Items") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineRenderDataTagOutShorthandUsesFieldByDefault(t *testing.T) {
	engine := New(
		WithDataDirective("article_list", func(call DataTagCall) (DataTagResult, error) {
			out, ok, err := call.Out()
			if err != nil {
				return DataTagResult{}, err
			}
			if !ok || out != "field" {
				return DataTagResult{}, fmt.Errorf("unexpected out: %q %v", out, ok)
			}

			return DataTagResult{
				Value: map[string]any{"title": "A"},
			}, nil
		}),
	)

	output, err := engine.RenderString(`@sq:article_list(out)<span>{{ field.title }}</span>`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<span>A</span>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderDataTagOutNilDefaultsToEmpty(t *testing.T) {
	engine := New(
		WithDataDirective("site", func(call DataTagCall) (DataTagResult, error) {
			return DataTagResult{}, nil
		}),
	)

	output, err := engine.RenderString(`@sq:site(out)@if(field)<h1>{{ field.title }}</h1>@else<span>empty</span>@endif`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<span>empty</span>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderDataTagBlockEmptyItemsRendersEmptyBranch(t *testing.T) {
	engine := New(
		WithDataDirective("arc", func(call DataTagCall) (DataTagResult, error) {
			return DataTagResult{Items: nil}, nil
		}),
	)

	output, err := engine.RenderString(`<ul>@sq:arc(pid=10)<li>{{ field.title }}</li>@empty<li>empty</li>@end</ul>`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<ul><li>empty</li></ul>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderExtendsAndComponent(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "layout.tpl", `<html><body>@yield('content')</body></html>`)
	writeTemplate(t, root, "badge.tpl", `<span class="badge">{{ slot }}</span>`)
	writeTemplate(t, root, "page.tpl", `@extends('layout')@section('content')@component('badge'){{ name }}-{{ count }}@endcomponent@endsection`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("page", map[string]any{
		"name":  "sq",
		"count": 2,
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := `<html><body><span class="badge">sq-2</span></body></html>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderEachComponentAndBlockDirective(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "chip.tpl", `<span class="chip">{{ slot }}</span>`)
	writeTemplate(t, root, "row.tpl", `@component('chip')@sq:box(){{ user.name }}:{{ user.index }}@end@endcomponent`)
	writeTemplate(t, root, "list.tpl", `<div>@each('row', users, 'user')</div>`)

	engine := New(
		WithFS(OSFS{Root: root}),
		WithBlockDirective("box", "", func(_ string, body string) (string, error) {
			return `<em>` + body + `</em>`, nil
		}),
	)

	output, err := engine.RenderFile("list", map[string]any{
		"users": []map[string]any{
			{"name": "A", "index": 1},
			{"name": "B", "index": 2},
		},
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := `<div><span class="chip"><em>A:1</em></span><span class="chip"><em>B:2</em></span></div>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderNamespacedTemplates(t *testing.T) {
	root := t.TempDir()
	cmsRoot := filepath.Join(root, "cms")

	writeTemplate(t, cmsRoot, "layouts/base.tpl", `<html><body>@yield('content')@include('cms:partials.footer')</body></html>`)
	writeTemplate(t, cmsRoot, "partials/footer.tpl", `<footer>{{ brand }}</footer>`)
	writeTemplate(t, cmsRoot, "components/badge.tpl", `<strong>{{ slot }}</strong>`)
	writeTemplate(t, cmsRoot, "rows/user.tpl", `<li>@component('cms:components.badge'){{ user.name }}@endcomponent</li>`)
	writeTemplate(t, cmsRoot, "rows/empty.tpl", `<p>empty</p>`)
	writeTemplate(t, root, "pages/home.tpl", `@extends('cms:layouts.base')@section('content')<ul>@each('cms:rows.user', users, 'user', 'cms:rows.empty')</ul>@endsection`)

	fs := OSFS{
		Root: root,
		Namespaces: map[string]string{
			"cms": cmsRoot,
		},
	}

	for _, name := range []string{
		"pages.home",
		"cms:layouts.base",
		"cms:partials.footer",
		"cms:components.badge",
		"cms:rows.user",
		"cms:rows.empty",
	} {
		if _, err := fs.Load(name); err != nil {
			t.Fatalf("Load(%q) error = %v", name, err)
		}
	}

	engine := New(WithFS(fs))

	output, err := engine.RenderFile("pages.home", map[string]any{
		"users": []map[string]any{
			{"name": "A"},
			{"name": "B"},
		},
		"brand": "Sparq",
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<html><body><ul><li><strong>A</strong></li><li><strong>B</strong></li></ul><footer>Sparq</footer></body></html>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderRealPageComposition(t *testing.T) {
	root := t.TempDir()
	cmsRoot := filepath.Join(root, "cms")

	writeTemplate(t, cmsRoot, "layouts/default.tpl", `<html><body>@yield('content')@include('cms:partials.footer', footer)</body></html>`)
	writeTemplate(t, cmsRoot, "partials/footer.tpl", `<footer>{{ text }}-{{ brand }}</footer>`)
	writeTemplate(t, cmsRoot, "components/card.tpl", `<article data-kind="{{ kind }}"><header>{{ title }}</header><section>{{ slot }}</section><aside>{{ aside }}</aside></article>`)
	writeTemplate(t, cmsRoot, "rows/article.tpl", `@component('cms:components.card', card)@sq:panel(){{ article.title }}:{{ card.kind }}@end@slot('aside'){{ article.author }}@endslot@endcomponent`)
	writeTemplate(t, cmsRoot, "rows/empty.tpl", `<p>empty</p>`)
	writeTemplate(t, root, "pages/dashboard.tpl", `@extends('cms:layouts.default')@section('content')<main>@include('partials.hero', hero)<ul>@each('cms:rows.article', articles, 'article', 'cms:rows.empty')</ul></main>@endsection`)
	writeTemplate(t, root, "partials/hero.tpl", `<h1>{{ title }}</h1><p>{{ brand }}</p>`)

	engine := New(
		WithFS(OSFS{
			Root: root,
			Namespaces: map[string]string{
				"cms": cmsRoot,
			},
		}),
		WithBlockDirective("panel", "", func(_ string, body string) (string, error) {
			return `<div class="panel">` + body + `</div>`, nil
		}),
	)

	output, err := engine.RenderFile("pages.dashboard", map[string]any{
		"brand": "Sparq",
		"hero": map[string]any{
			"title": "Dashboard",
		},
		"footer": map[string]any{
			"text": "Powered",
		},
		"card": map[string]any{
			"kind":  "story",
			"title": "Featured",
		},
		"articles": []map[string]any{
			{"title": "First", "author": "Alice"},
			{"title": "Second", "author": "Bob"},
		},
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := `<html><body><main><h1>Dashboard</h1><p>Sparq</p><ul><article data-kind="story"><header>Featured</header><section><div class="panel">First:story</div></section><aside>Alice</aside></article><article data-kind="story"><header>Featured</header><section><div class="panel">Second:story</div></section><aside>Bob</aside></article></ul></main><footer>Powered-Sparq</footer></body></html>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderNamespacedTemplateNotFound(t *testing.T) {
	root := t.TempDir()

	engine := New(WithFS(OSFS{
		Root: root,
		Namespaces: map[string]string{
			"cms": filepath.Join(root, "cms"),
		},
	}))

	_, err := engine.RenderFile("blog:home", nil)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}

func TestEngineRenderDottedTemplateNames(t *testing.T) {
	root := t.TempDir()

	writeTemplate(t, root, "default/view.tpl", `<html><body>@yield('content')</body></html>`)
	writeTemplate(t, root, "pages/home.tpl", `@extends('default.view')@section('content')<p>{{ title }}</p>@endsection`)

	engine := New(WithFS(OSFS{Root: root}))

	output, err := engine.RenderFile("pages.home", map[string]any{
		"title": "Blade",
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<html><body><p>Blade</p></body></html>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderUnclosedCustomBlockDirectiveError(t *testing.T) {
	engine := New(
		WithBlockDirective("panel", "", func(_ string, body string) (string, error) {
			return `<section>` + body + `</section>`, nil
		}),
	)

	_, err := engine.RenderString(`@sq:panel()Hello`, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "自定义指令未正常闭合: @sq:panel") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineRenderFileErrorIncludesSourceContext(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "page.tpl", "<section>\n@component('card')Hello")

	engine := New(WithFS(OSFS{Root: root}))

	_, err := engine.RenderFile("page", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	message := err.Error()
	for _, expected := range []string{
		"page.tpl:2:1",
		"component 未正常闭合",
		"@component('card')Hello",
		"^",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected %q in error, got %v", expected, err)
		}
	}
}

func TestEngineRenderFileErrorIncludesTemplateChain(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "page.tpl", `@include('partial')`)
	writeTemplate(t, root, "partial.tpl", `@include('missing')`)

	engine := New(WithFS(OSFS{Root: root}))

	_, err := engine.RenderFile("page", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	message := err.Error()
	for _, expected := range []string{
		"模板渲染失败 [page -> partial -> missing]",
		"file does not exist",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected %q in error, got %v", expected, err)
		}
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected errors.Is(err, os.ErrNotExist), got %v", err)
	}
}

func TestEngineRenderUnless(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(`@unless(hidden)<span>{{ text }}</span>@endunless`, map[string]any{
		"hidden": false,
		"text":   "ok",
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "<span>ok</span>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderLogicalComparisonExpressions(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@if((count >= 2 && !hidden) || role == 'admin')<b>{{ name }}</b>@elseif(count == 1)<i>single</i>@else<span>empty</span>@endif`,
		map[string]any{
			"count":  2,
			"hidden": false,
			"role":   "member",
			"name":   "ok",
		},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "<b>ok</b>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderElseIfComparison(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@if(count > 2)many@elseif(count == 2 && role != 'iuest')pair@else other@endif`,
		map[string]any{
			"count": 2,
			"role":  "member",
		},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "pair"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderForLoop(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@for(i in 1..3)<span>{{ i }}:{{ loop.index }}</span>@endfor`,
		nil,
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "<span>1:0</span><span>2:1</span><span>3:2</span>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderForLoopDescending(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@for(i in 3..1)<span>{{ i }}</span>@endfor`,
		nil,
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "<span>3</span><span>2</span><span>1</span>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderListWithEmpty(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@list(items, as=item)<span>{{ item.name }}</span>@empty<em>empty</em>@endlist`,
		map[string]any{"items": []map[string]any{}},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "<em>empty</em>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderChildrenWithWrapAndEmpty(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@list(items, as=item)<section>{{ item.title }}@children(as=child, wrap="ul", class="submenu")<li>{{ child.title }}</li>@empty<span>none</span>@endchildren</section>@endlist`,
		map[string]any{
			"items": []map[string]any{
				{
					"title": "A",
					"children": []map[string]any{
						{"title": "A1"},
						{"title": "A2"},
					},
				},
				{
					"title":    "B",
					"children": []map[string]any{},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<section>A<ul class="submenu"><li>A1</li><li>A2</li></ul></section><section>B<span>none</span></section>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderChildrenWrapClassEscaped(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@list(items, as=item)@children(as=child, wrap="div", class="a & b")<span>{{ child.title }}</span>@endchildren@endlist`,
		map[string]any{
			"items": []map[string]any{
				{
					"children": []map[string]any{
						{"title": "A"},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<div class="a &amp; b"><span>A</span></div>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderChildrenFieldOverride(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@list(items, as=item){{ item.title }}@children(field=nodes, as=node, wrap="div")<b>{{ node.title }}</b>@endchildren@endlist`,
		map[string]any{
			"items": []map[string]any{
				{
					"title": "A",
					"nodes": []map[string]any{
						{"title": "A1"},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "A<div><b>A1</b></div>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderChildrenInsideDataTag(t *testing.T) {
	engine := New(
		WithDataDirective("nav", func(call DataTagCall) (DataTagResult, error) {
			return Items(
				map[string]any{
					"title": "Root",
					"children": []map[string]any{
						{"title": "Child"},
					},
				},
			), nil
		}),
	)

	output, err := engine.RenderString(
		`@sq:nav(as=item){{ item.title }}@children(as=child, wrap="ul")<li>{{ child.title }}</li>@endchildren@end`,
		nil,
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "Root<ul><li>Child</li></ul>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderChildrenRequiresCurrentItem(t *testing.T) {
	engine := New()

	_, err := engine.RenderString(`@children(as=child)<span>{{ child.title }}</span>@endchildren`, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "@children 只能在列表项上下文中使用") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineRenderFlowDirectivesWithoutBrandPrefix(t *testing.T) {
	engine := New()

	output, err := engine.RenderString(
		`@if(show)@list(items, as=item){{ item.title }}@children(as=child, wrap="ul", class="submenu")<li>{{ child.title }}</li>@empty none@end@empty empty@end@else hidden@end`,
		map[string]any{
			"show": true,
			"items": []map[string]any{
				{
					"title": "A",
					"children": []map[string]any{
						{"title": "A1"},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `A<ul class="submenu"><li>A1</li></ul>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderDataTagDirectiveWithGenericEnd(t *testing.T) {
	engine := New(
		WithDataDirective("arc", func(call DataTagCall) (DataTagResult, error) {
			return Items(map[string]any{"title": "A"}), nil
		}),
	)

	output, err := engine.RenderString(`@sq:arc(as=article)<b>{{ article.title }}</b>@end`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<b>A</b>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestEngineRenderWhileUnsupported(t *testing.T) {
	engine := New()

	_, err := engine.RenderString(`@while(true)x@endwhile`, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "当前版本尚未支持该指令: @while") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeTemplate(t *testing.T, root, name, content string) {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(name), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", name, err)
	}
}

type errorFS struct {
	err error
}

func (fs errorFS) Load(string) (*TemplateSource, error) {
	return nil, fs.err
}
