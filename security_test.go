package blade_test

import (
	"html/template"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	blade "github.com/zane-xin/sparq-blade"
)

func TestRenderRejectsNonIterableValue(t *testing.T) {
	engine := blade.New()

	_, err := engine.RenderString(`@foreach(items as item){{ item }}@endforeach`, map[string]any{"items": 42})
	if err == nil || !strings.Contains(err.Error(), "foreach 不支持遍历 int") {
		t.Fatalf("error = %v, want non-iterable error", err)
	}
}

func TestRenderRejectsOversizedCollection(t *testing.T) {
	engine := blade.New()
	items := make([]int, 10001)

	_, err := engine.RenderString(`@foreach(items as item){{ item }}@endforeach`, map[string]any{"items": items})
	if err == nil || !strings.Contains(err.Error(), "集合循环超过最大迭代次数 10000") {
		t.Fatalf("error = %v, want collection limit error", err)
	}
}

func TestRenderRejectsOverflowingForRange(t *testing.T) {
	engine := blade.New()
	tests := []struct {
		name  string
		start int64
		end   int64
	}{
		{name: "ascending", start: math.MinInt64, end: math.MaxInt64},
		{name: "descending", start: math.MaxInt64, end: math.MinInt64},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := engine.RenderString(
				`@for(i in start..end){{ i }}@endfor`,
				map[string]any{"start": test.start, "end": test.end},
			)
			if err == nil || !strings.Contains(err.Error(), "for 循环超过最大迭代次数 10000") {
				t.Fatalf("error = %v, want for limit error", err)
			}
		})
	}
}

func TestRenderRejectsTrustedTemplatePointer(t *testing.T) {
	trusted := template.HTML(`<script>alert("unsafe")</script>`)
	engine := blade.New(blade.WithFunction("plugin.html", func(...any) (any, error) {
		return &trusted, nil
	}))

	_, err := engine.RenderString(`{{ plugin.html() }}`, nil)
	if err == nil || !strings.Contains(err.Error(), "不允许返回可信模板内容") {
		t.Fatalf("error = %v, want trusted content error", err)
	}
}

func TestRenderRejectsInheritanceCycle(t *testing.T) {
	root := t.TempDir()
	writeSecurityTemplate(t, root, "a.tpl", `@extends('b')`)
	writeSecurityTemplate(t, root, "b.tpl", `@extends('a')`)

	engine := blade.New(blade.WithFS(blade.OSFS{Root: root}))
	_, err := engine.RenderFile("a", nil)
	if err == nil || !strings.Contains(err.Error(), "模板继承存在循环调用: a.tpl -> b.tpl -> a.tpl") {
		t.Fatalf("error = %v, want inheritance cycle error", err)
	}
}

func TestRenderRejectsRuntimeCompositionCycles(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "include", content: `@include('loop')`},
		{name: "component", content: `@component('loop')content@endcomponent`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeSecurityTemplate(t, root, "loop.tpl", test.content)

			engine := blade.New(blade.WithFS(blade.OSFS{Root: root}))
			_, err := engine.RenderFile("loop", nil)
			if err == nil || !strings.Contains(err.Error(), "模板组合存在循环调用: loop.tpl -> loop.tpl") {
				t.Fatalf("error = %v, want composition cycle error", err)
			}
		})
	}
}

func TestRenderRejectsExcessiveCompositionDepth(t *testing.T) {
	root := t.TempDir()
	for index := 0; index <= 64; index++ {
		content := "done"
		if index < 64 {
			content = `@include('` + strconv.Itoa(index+1) + `')`
		}
		writeSecurityTemplate(t, root, strconv.Itoa(index)+".tpl", content)
	}

	engine := blade.New(blade.WithFS(blade.OSFS{Root: root}))
	_, err := engine.RenderFile("0", nil)
	if err == nil || !strings.Contains(err.Error(), "模板组合超过最大深度 64") {
		t.Fatalf("error = %v, want composition depth error", err)
	}
}

func TestRenderRejectsOversizedDataTagResult(t *testing.T) {
	items := make([]any, 10001)
	engine := blade.New(blade.WithDataDirective("arc", func(blade.DataTagCall) (blade.DataTagResult, error) {
		return blade.Items(items...), nil
	}))

	_, err := engine.RenderString(`@sq:arc(as=item){{ item }}@end`, nil)
	if err == nil || !strings.Contains(err.Error(), "@sq:arc 返回项数超过最大迭代次数 10000") {
		t.Fatalf("error = %v, want data tag limit error", err)
	}
}

func TestCompositionTrackingPreservesIncludeScope(t *testing.T) {
	root := t.TempDir()
	writeSecurityTemplate(t, root, "page.tpl", `@include('set'){{ shared }}`)
	writeSecurityTemplate(t, root, "set.tpl", `@sq:site(out='shared')`)

	engine := blade.New(
		blade.WithFS(blade.OSFS{Root: root}),
		blade.WithDataDirective("site", func(blade.DataTagCall) (blade.DataTagResult, error) {
			return blade.ValueOf("ok"), nil
		}),
	)
	output, err := engine.RenderFile("page", nil)
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}
	if output != "ok" {
		t.Fatalf("output = %q, want ok", output)
	}
}

func writeSecurityTemplate(t *testing.T, root string, name string, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(name), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", name, err)
	}
}
