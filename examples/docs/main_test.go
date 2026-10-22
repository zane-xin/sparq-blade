package main

import (
	"strings"
	"testing"

	blade "github.com/zane-xin/sparq-blade"
)

func TestDocsPagesRender(t *testing.T) {
	renderer := blade.New(
		blade.WithFS(blade.OSFS{Root: viewsRoot()}),
		blade.WithTag("title", renderTitle),
		blade.WithTag("seo", renderSEO),
		blade.WithDataDirective("nav", renderNav),
		blade.WithDataDirective("scenarios", renderScenarios),
		blade.WithDataDirective("steps", renderSteps),
		blade.WithDataDirective("arc", renderArticles),
	)

	for _, item := range navItems {
		item := item
		t.Run(item.Template, func(t *testing.T) {
			output, err := renderer.RenderFile(item.Template, ViewContext{
				Site:  SiteContext{Name: "Blade 示例操作文档"},
				Route: RouteContext{Path: item.URL},
				Page:  pageData(item),
				Meta:  item.Meta,
			})
			if err != nil {
				t.Fatalf("RenderFile(%q) error = %v", item.Template, err)
			}
			if !strings.Contains(output, item.Meta.Title) {
				t.Fatalf("rendered page %q does not contain title %q", item.Template, item.Meta.Title)
			}
		})
	}
}

func TestDocsIncludeAndComponentDataRender(t *testing.T) {
	renderer := blade.New(
		blade.WithFS(blade.OSFS{Root: viewsRoot()}),
		blade.WithTag("title", renderTitle),
		blade.WithTag("seo", renderSEO),
		blade.WithDataDirective("nav", renderNav),
		blade.WithDataDirective("steps", renderSteps),
	)

	basic := renderDocsPage(t, renderer, navItems[1])
	for _, want := range []string{"include 传入的标题", "当前页面：基础渲染"} {
		if !strings.Contains(basic, want) {
			t.Fatalf("basic page missing %q", want)
		}
	}

	components := renderDocsPage(t, renderer, navItems[2])
	for _, want := range []string{"component 传入的标题", "组件标题来自传入数据"} {
		if !strings.Contains(components, want) {
			t.Fatalf("components page missing %q", want)
		}
	}
}

func renderDocsPage(t *testing.T, renderer *blade.Engine, item navItem) string {
	t.Helper()

	output, err := renderer.RenderFile(item.Template, ViewContext{
		Site:  SiteContext{Name: "Blade 示例操作文档"},
		Route: RouteContext{Path: item.URL},
		Page:  pageData(item),
		Meta:  item.Meta,
	})
	if err != nil {
		t.Fatalf("RenderFile(%q) error = %v", item.Template, err)
	}

	return output
}
