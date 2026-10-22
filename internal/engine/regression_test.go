package engine

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zane-xin/sparq-blade/internal/compiler"
	"github.com/zane-xin/sparq-blade/internal/directives"
)

func TestEngineRejectsNilDirectiveHandlersDuringInitialization(t *testing.T) {
	tests := []struct {
		name   string
		option Option
	}{
		{name: "single directive", option: WithDirective("badge", nil)},
		{name: "block directive", option: WithBlockDirective("panel", "", nil)},
		{name: "output tag", option: WithTag("pager", nil)},
		{name: "data directive", option: WithDataDirective("arc", nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					t.Fatal("expected initialization panic, got nil")
				}
				if !strings.Contains(fmt.Sprint(recovered), "不能为空") {
					t.Fatalf("unexpected panic: %v", recovered)
				}
			}()

			New(test.option)
		})
	}
}

func TestEngineUsesDirectiveRegistrySnapshot(t *testing.T) {
	registry := directives.NewRegistry()
	registry.Register("badge", func(expression string) (string, error) {
		return fmt.Sprintf(`<strong>{{ bladeEcho . %q }}</strong>`, expression), nil
	})

	engine := New(WithDirectives(registry))
	registry.Register("later", func(string) (string, error) {
		return "late", nil
	})

	output, err := engine.RenderString(`@sq:badge(name)`, map[string]any{"name": "Blade"})
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}
	if output != "<strong>Blade</strong>" {
		t.Fatalf("unexpected output: %q", output)
	}

	_, err = engine.RenderString(`@sq:later()`, nil)
	if err == nil || !strings.Contains(err.Error(), "当前版本尚未支持该指令: @sq:later") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineRejectsExplicitBusinessDirectiveNamespaces(t *testing.T) {
	tests := []struct {
		name   string
		option Option
	}{
		{name: "single directive", option: WithDirective("pt:badge", func(string) (string, error) { return "", nil })},
		{name: "block directive", option: WithBlockDirective("custom:panel", "", func(string, string) (string, error) { return "", nil })},
		{name: "output tag", option: WithTag("sq:pager", func(TagCall) (template.HTML, error) { return "", nil })},
		{name: "data directive", option: WithDataDirective("other:arc", func(DataTagCall) (DataTagResult, error) { return EmptyResult(), nil })},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					t.Fatal("expected initialization panic, got nil")
				}
				if !strings.Contains(fmt.Sprint(recovered), "业务指令注册只接受局部名称") {
					t.Fatalf("unexpected panic: %v", recovered)
				}
			}()

			New(test.option)
		})
	}
}

func TestEngineOutputTagAllowsAsParameterInSingleMode(t *testing.T) {
	engine := New(WithTag("formatter", func(call TagCall) (template.HTML, error) {
		format, err := call.GetString("as")
		if err != nil {
			return "", err
		}
		return template.HTML(format), nil
	}))

	output, err := engine.RenderString(`@sq:formatter(as="date")`, nil)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}
	if output != "date" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestEngineInheritanceResolvesSectionsInsideBusinessBlocks(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "layout.tpl", `@sq:panel()@yield('content')@end`)
	writeTemplate(t, root, "page.tpl", `@extends('layout')@section('content')child@endsection`)

	engine := New(
		WithFS(OSFS{Root: root}),
		WithBlockDirective("panel", "", func(_ string, body string) (string, error) {
			return "<section>" + body + "</section>", nil
		}),
	)

	output, err := engine.RenderFile("page", nil)
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}
	if output != "<section>child</section>" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestEngineInheritanceResolvesSectionsInsideBusinessEmptyBranch(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "layout.tpl", `@sq:arc()@yield('content')@empty@yield('empty')@end`)
	writeTemplate(t, root, "page.tpl", `@extends('layout')@section('content')content@endsection@section('empty')empty@endsection`)

	engine := New(
		WithFS(OSFS{Root: root}),
		WithDataDirective("arc", func(DataTagCall) (DataTagResult, error) {
			return EmptyResult(), nil
		}),
	)

	output, err := engine.RenderFile("page", nil)
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}
	if output != "empty" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestEngineParseFileReturnsIndependentAST(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root, "page.tpl", "original")

	engine := New(WithFS(OSFS{Root: root}))
	document, err := engine.ParseFile("page")
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	document.Nodes[0].(*compiler.TextNode).Text = "changed"

	output, err := engine.RenderFile("page", nil)
	if err != nil {
		t.Fatalf("RenderFile() error = %v", err)
	}
	if output != "original" {
		t.Fatalf("cached AST was modified through ParseFile: %q", output)
	}
}

func TestOSFSVersionTracksContentWhenMetadataIsUnchanged(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "page.tpl")
	if err := os.WriteFile(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("write first template: %v", err)
	}

	fs := OSFS{Root: root}
	first, err := fs.Load("page")
	if err != nil {
		t.Fatalf("first Load() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat template: %v", err)
	}

	if err := os.WriteFile(path, []byte("other"), 0o644); err != nil {
		t.Fatalf("write second template: %v", err)
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatalf("restore template timestamp: %v", err)
	}

	second, err := fs.Load("page")
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}
	if string(second.Content) != "other" {
		t.Fatalf("unexpected cached content: %q", second.Content)
	}
	if first.Version == second.Version {
		t.Fatalf("content change did not invalidate version: %s", first.Version)
	}
}

func TestEngineRejectsNilDirectiveRegistryDuringInitialization(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected initialization panic, got nil")
		}
		if !strings.Contains(fmt.Sprint(recovered), "自定义指令注册表不能为空") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	New(WithDirectives(nil))
}

func TestEngineRejectsBuiltinNamesFromDirectiveRegistry(t *testing.T) {
	registry := directives.NewRegistry()
	registry.Register("if", func(string) (string, error) {
		return "", nil
	})

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected initialization panic, got nil")
		}
		if !strings.Contains(fmt.Sprint(recovered), "业务指令名与内置指令冲突: @sq:if") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	New(WithDirectives(registry))
}
