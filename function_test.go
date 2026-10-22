package blade_test

import (
	"fmt"
	"html/template"
	"strings"
	"testing"
	"time"

	blade "github.com/zane-xin/sparq-blade"
)

func TestBuiltinFunctions(t *testing.T) {
	engine := blade.New()
	createdAt := time.Date(2026, time.July, 30, 15, 4, 5, 0, time.FixedZone("CST", 8*60*60))
	output, err := engine.RenderString(
		`{{ upper(user.name) }}|{{ lower("SPARQ") }}|{{ len("模板") }}|{{ default(empty, "fallback") }}|{{ coalesce(missing, user.name) }}|{{ formatDate(createdAt, "2006-01-02 15:04") }}@if(contains(user.roles, "admin"))|admin@endif`,
		map[string]any{
			"user":      map[string]any{"name": "Sparq", "roles": []string{"editor", "admin"}},
			"empty":     "",
			"missing":   nil,
			"createdAt": createdAt,
		},
	)
	if err != nil {
		t.Fatalf("render built-in functions: %v", err)
	}
	if output != "SPARQ|sparq|2|fallback|Sparq|2026-07-30 15:04|admin" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestAssetFunctionUsesConfiguredStaticPrefix(t *testing.T) {
	engine := blade.New(blade.WithAssetPrefix("/static/site"))
	output, err := engine.RenderString(`<link href="{{ asset("css/app.css?v=2#theme") }}">`, nil)
	if err != nil {
		t.Fatalf("render asset: %v", err)
	}
	if output != `<link href="/static/site/css/app.css?v=2#theme">` {
		t.Fatalf("unexpected asset output: %q", output)
	}
}

func TestCustomNamespacedFunctionSupportsNestedCalls(t *testing.T) {
	engine := blade.New(
		blade.WithFunction("cms.label", func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("label expects one argument")
			}
			return "[" + fmt.Sprint(args[0]) + "]", nil
		}),
	)
	output, err := engine.RenderString(`{{ cms.label(upper(article.title)) }}`, map[string]any{
		"article": map[string]any{"title": "news"},
	})
	if err != nil {
		t.Fatalf("render custom function: %v", err)
	}
	if output != "[NEWS]" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestFunctionWorksInBusinessTagParams(t *testing.T) {
	engine := blade.New(
		blade.WithTag("label", func(call blade.TagCall) (template.HTML, error) {
			value, err := call.GetString("value")
			if err != nil {
				return "", err
			}
			return template.HTML(value), nil
		}),
	)
	output, err := engine.RenderString(`@sq:label(value=upper(user.name))`, map[string]any{
		"user": map[string]any{"name": "sparq"},
	})
	if err != nil {
		t.Fatalf("render business tag function argument: %v", err)
	}
	if output != "SPARQ" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestFunctionErrorsAreExplicit(t *testing.T) {
	tests := []struct {
		name    string
		engine  *blade.Engine
		source  string
		message string
	}{
		{name: "unknown", engine: blade.New(), source: `{{ missing(value) }}`, message: "未注册模板函数: missing"},
		{name: "invalid args", engine: blade.New(), source: `{{ upper("a", "b") }}`, message: "upper 需要 1 个参数"},
		{name: "asset traversal", engine: blade.New(), source: `{{ asset("../secret") }}`, message: "asset 路径不能包含上级目录"},
		{name: "trusted html", engine: blade.New(blade.WithFunction("plugin.html", func(...any) (any, error) {
			return template.HTML("<strong>unsafe</strong>"), nil
		})), source: `{{ plugin.html() }}`, message: "不允许返回可信模板内容"},
		{name: "trusted url", engine: blade.New(blade.WithFunction("plugin.url", func(...any) (any, error) {
			return template.URL("javascrisq:unsafe"), nil
		})), source: `{{ plugin.url() }}`, message: "不允许返回可信模板内容"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.engine.RenderString(test.source, map[string]any{"value": "x"})
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("error = %v, want message %q", err, test.message)
			}
		})
	}
}
