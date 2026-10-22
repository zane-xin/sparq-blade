package blade_test

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	blade "github.com/zane-xin/sparq-blade"
)

func TestRenderStringPublicAPI(t *testing.T) {
	output, err := blade.RenderString(`<h1>{{ title }}</h1>`, map[string]any{
		"title": "Blade",
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "<h1>Blade</h1>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestRenderStringRejectsDollarVariablePrefix(t *testing.T) {
	renderer := blade.New()
	testCases := []struct {
		name   string
		source string
	}{
		{name: "输出表达式", source: `{{ $name }}`},
		{name: "条件表达式", source: `@if($enabled)yes@endif`},
		{name: "循环表达式", source: `@foreach($items as item){{ item }}@endforeach`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := renderer.RenderString(testCase.source, map[string]any{
				"name":    "Sparq",
				"enabled": true,
				"items":   []string{"A"},
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "模板变量不再支持 $ 前缀") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEngineDoesNotRegisterLegacyPTBrandAlias(t *testing.T) {
	renderer := blade.New(
		blade.WithDataDirective("arc", func(blade.DataTagCall) (blade.DataTagResult, error) {
			return blade.EmptyResult(), nil
		}),
	)
	testCases := []struct {
		name      string
		source    string
		directive string
	}{
		{name: "旧品牌流程指令", source: `@pt:if(true)yes@pt:end`, directive: "@pt:if"},
		{name: "旧品牌数据标签", source: `@pt:arc(out)`, directive: "@pt:arc"},
		{name: "品牌条件指令", source: `@sq:if(true)yes@end`, directive: "@sq:if"},
		{name: "品牌列表指令", source: `@sq:list(items, as=item){{ item }}@end`, directive: "@sq:list"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := renderer.RenderString(testCase.source, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), testCase.directive) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDataTagRejectsDollarOutputName(t *testing.T) {
	renderer := blade.New(
		blade.WithDataDirective("arc", func(blade.DataTagCall) (blade.DataTagResult, error) {
			return blade.EmptyResult(), nil
		}),
	)

	_, err := renderer.RenderString(`@sq:arc(out='$items')`, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "参数 out 不再支持 $ 前缀") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewWithDirectivePublicAPI(t *testing.T) {
	renderer := blade.New(
		blade.WithDirective("badge", func(expression string) (string, error) {
			return fmt.Sprintf(`<strong>{{ bladeEcho . %q }}</strong>`, expression), nil
		}),
	)

	output, err := renderer.RenderString(`@sq:badge(name)`, map[string]any{
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

func TestNewWithBlockDirectivePublicAPI(t *testing.T) {
	renderer := blade.New(
		blade.WithBlockDirective("panel", "", func(_ string, body string) (string, error) {
			return `<section class="panel">` + body + `</section>`, nil
		}),
	)

	output, err := renderer.RenderString(`@sq:panel()Hello {{ name }}@end`, map[string]any{
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

func TestNewWithFSPublicAPI(t *testing.T) {
	root := t.TempDir()
	writePublicTemplate(t, root, "layout.tpl", `<html><body>@yield('content')</body></html>`)
	writePublicTemplate(t, root, "page.tpl", `@extends('layout')@section('content')<p>{{ name }}-{{ count }}</p>@endsection`)

	renderer := blade.New(
		blade.WithFS(blade.OSFS{Root: root}),
	)

	output, err := renderer.RenderFile("page", map[string]any{
		"name":  "sq",
		"count": 2,
	})
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}

	expected := "<html><body><p>sq-2</p></body></html>"
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestParsePublicAPIErrorIncludesSourceContext(t *testing.T) {
	_, err := blade.Parse("<main>\n@component('card')Hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	message := err.Error()
	for _, expected := range []string{
		"component 未正常闭合",
		"@component('card')Hello",
		"^",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected %q in error, got %v", expected, err)
		}
	}
}

func TestNewWithTagPublicAPI(t *testing.T) {
	renderer := blade.New(
		blade.WithTag("snippet", func(call blade.TagCall) (template.HTML, error) {
			return template.HTML(fmt.Sprintf(`<span>%s:%v</span>`, call.Target, call.Params["limit"])), nil
		}),
	)

	output, err := renderer.RenderString(`@sq:snippet('cms:arc', limit=3)`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<span>cms:arc:3</span>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestNewWithTagDoesNotRegisterBareAlias(t *testing.T) {
	renderer := blade.New(
		blade.WithTag("snippet", func(blade.TagCall) (template.HTML, error) {
			return "", nil
		}),
	)

	_, err := renderer.RenderString(`@snippet()`, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "@snippet") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestNewWithTagContextPublicAPI(t *testing.T) {
	type viewContext struct {
		Title string
	}

	renderer := blade.New(
		blade.WithTag("title", func(call blade.TagCall) (template.HTML, error) {
			ctx, ok := call.Context().(viewContext)
			if !ok {
				return "", fmt.Errorf("unexpected context: %T", call.Context())
			}

			return template.HTML(template.HTMLEscapeString(ctx.Title)), nil
		}),
	)

	output, err := renderer.RenderString(`<title>@sq:title("page")</title>`, viewContext{
		Title: "News",
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<title>News</title>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestNewWithDataTagPublicAPI(t *testing.T) {
	renderer := blade.New(
		blade.WithDataDirective("arc", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			limit, err := call.GetInt("limit")
			if err != nil {
				return blade.DataTagResult{}, err
			}

			return blade.DataTagResult{
				Items: []any{
					map[string]any{"title": fmt.Sprintf("item-%d", limit)},
				},
			}, nil
		}),
	)

	output, err := renderer.RenderString(`@sq:arc(limit=3, out)@foreach(field as item)<span>{{ item.title }}</span>@endforeach`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<span>item-3</span>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestNewWithDataTagContextPublicAPI(t *testing.T) {
	type category struct {
		ID int
	}
	type viewContext struct {
		CurrentCategory category
	}

	renderer := blade.New(
		blade.WithDataDirective("arc", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			cid, err := call.GetIntDefault("cid", 0)
			if err != nil {
				return blade.DataTagResult{}, err
			}
			if cid == 0 {
				ctx, ok := call.Context().(viewContext)
				if !ok {
					return blade.DataTagResult{}, fmt.Errorf("unexpected context: %T", call.Context())
				}
				cid = ctx.CurrentCategory.ID
			}

			return blade.Items(
				blade.Item(blade.Fields().Set("title", fmt.Sprintf("category-%d", cid))),
			), nil
		}),
	)

	output, err := renderer.RenderString(`@sq:arc(as=field)<span>{{ field.title }}</span>@end`, viewContext{
		CurrentCategory: category{ID: 7},
	})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<span>category-7</span>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestNewWithDataTagValuePublicAPI(t *testing.T) {
	renderer := blade.New(
		blade.WithDataDirective("site", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			return blade.ValueOf(
				blade.Fields().
					Set("title", "SQ CMS"),
			), nil
		}),
	)

	output, err := renderer.RenderString(`@sq:site(out)<h1>{{ field.title }}</h1>`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<h1>SQ CMS</h1>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestDataTagRowsAndObjectPublicAPI(t *testing.T) {
	type article struct {
		Title string
		URL   string
	}
	type site struct {
		Title string
		URL   string
	}

	renderer := blade.New(
		blade.WithDataDirective("articles", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			return blade.Rows([]article{
				{Title: "A", URL: "/a"},
				{Title: "B", URL: "/b"},
			}), nil
		}),
		blade.WithDataDirective("links", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			return blade.Rows(
				map[string]any{"title": "A", "url": "/a"},
				map[string]any{"title": "B", "url": "/b"},
			), nil
		}),
		blade.WithDataDirective("site", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			return blade.Object(site{Title: "SQ CMS", URL: "/"}), nil
		}),
	)

	output, err := renderer.RenderString(
		`@sq:articles(as=article)<a href="{{ article.URL }}">{{ article.Title }}</a>@end`+
			`@sq:links(as=link)<a href="{{ link.url }}">{{ link.title }}</a>@end`+
			`@sq:site(out=site)<strong>{{ site.Title }}</strong>`,
		nil,
	)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := `<a href="/a">A</a><a href="/b">B</a><a href="/a">A</a><a href="/b">B</a><strong>SQ CMS</strong>`
	if output != expected {
		t.Fatalf("unexpected output: got %q want %q", output, expected)
	}
}

func TestDataTagHelperPublicAPI(t *testing.T) {
	fields := blade.Fields().
		Set("id", 1).
		Set("title", "A").
		Merge(map[string]any{"url": "/a"})

	item := blade.Item(fields).(map[string]any)
	if item["id"] != 1 || item["title"] != "A" || item["url"] != "/a" {
		t.Fatalf("unexpected item: %#v", item)
	}

	list := blade.Items(blade.Item(blade.Fields().Set("title", "A")))
	if len(list.Items) != 1 {
		t.Fatalf("unexpected items len: %d", len(list.Items))
	}

	value := blade.ValueOf(blade.Fields().Set("title", "Site"))
	if value.Value == nil {
		t.Fatal("expected value, got nil")
	}

	rows := blade.Rows([]map[string]any{{"title": "A"}, {"title": "B"}})
	if len(rows.Items) != 2 {
		t.Fatalf("unexpected rows len: %d", len(rows.Items))
	}

	object := blade.Object(map[string]any{"title": "Site"})
	if object.Value == nil {
		t.Fatal("expected object, got nil")
	}

	empty := blade.EmptyResult()
	if empty.Value != nil || empty.Items != nil {
		t.Fatalf("unexpected empty result: %#v", empty)
	}
}

func writePublicTemplate(t *testing.T, root, name, content string) {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", name, err)
	}
}
