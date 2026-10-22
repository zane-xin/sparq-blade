package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	blade "github.com/zane-xin/sparq-blade"
)

type ViewContext struct {
	Site  SiteContext
	Route RouteContext
	Page  PageContext
	Crumb []Crumb
	Docs  []DocItem
}

type SiteContext struct {
	Name string
	URL  string
}

type RouteContext struct {
	Path string
}

type PageContext struct {
	Title      string
	Summary    string
	CategoryID int
	Pager      Pager
	Items      []any
	Include    map[string]any
	Card       map[string]any
}

type Pager struct {
	Page  int
	Size  int
	Total int
}

type Crumb struct {
	Title string
	URL   string
}

type DocItem struct {
	Title string
	Body  string
}

type pageSpec struct {
	Path       string
	View       string
	Title      string
	Summary    string
	CategoryID int
	Pager      Pager
	Docs       []DocItem
}

var pages = []pageSpec{
	{
		Path:       "/",
		View:       "home",
		Title:      "CMS 模板示例",
		Summary:    "一个小型 HTTP 示例，用页面文档的方式演示模板指令、数据标签、模板组合和自定义扩展。",
		CategoryID: 7,
		Pager:      Pager{Page: 2, Size: 2, Total: 23},
		Docs: []DocItem{
			{Title: "基础指令", Body: "覆盖输出、条件、列表、空状态、计数循环和高级遍历。"},
			{Title: "模板组合", Body: "覆盖布局继承、section/yield、include、component、slot 和 each。"},
			{Title: "业务数据", Body: `覆盖 WithDataDirective、WithTag、上下文推导、out、Rows/Object/EmptyResult。`},
			{Title: "模板查找层级", Body: "通过 LayeredFS 让主题模板覆盖应用模板和模块默认模板。"},
		},
	},
	{
		Path:       "/directives",
		View:       "directives",
		Title:      "基础指令",
		Summary:    "演示模板输出、条件判断、列表空状态、计数循环和保留给高阶场景的遍历写法。",
		CategoryID: 7,
		Pager:      Pager{Page: 1, Size: 10, Total: 10},
	},
	{
		Path:       "/templates",
		View:       "templates",
		Title:      "模板组合",
		Summary:    "演示布局继承、局部模板、组件、插槽、each 和命名空间模板查找。",
		CategoryID: 7,
		Pager:      Pager{Page: 1, Size: 10, Total: 10},
	},
	{
		Path:       "/data",
		View:       "data",
		Title:      "数据处理",
		Summary:    "演示数据型业务指令返回列表、单对象、复合对象和空结果时的模板写法。",
		CategoryID: 7,
		Pager:      Pager{Page: 1, Size: 2, Total: 42},
	},
	{
		Path:       "/articles",
		View:       "articles",
		Title:      "文章列表",
		Summary:    "演示 @sq:arc 的上下文推导，以及显式参数覆盖上下文。",
		CategoryID: 7,
		Pager:      Pager{Page: 2, Size: 2, Total: 23},
	},
	{
		Path:       "/navigation",
		View:       "navigation",
		Title:      "导航子级",
		Summary:    "演示 @children 渲染当前项的直接子级，并按需输出包裹标签。",
		CategoryID: 0,
		Pager:      Pager{Page: 1, Size: 10, Total: 10},
	},
	{
		Path:       "/pagination",
		View:       "pagination",
		Title:      "分页与模板覆盖",
		Summary:    "演示 @sq:pager 输出分页，并通过 LayeredFS 使用主题模板覆盖模块默认模板。",
		CategoryID: 7,
		Pager:      Pager{Page: 3, Size: 2, Total: 31},
	},
	{
		Path:       "/custom",
		View:       "custom",
		Title:      "自定义扩展",
		Summary:    "演示单指令、块级指令、输出型业务标签和数据型业务指令在 Go 侧的注册方式。",
		CategoryID: 7,
		Pager:      Pager{Page: 1, Size: 10, Total: 10},
	},
}

func main() {
	addr := "127.0.0.1:8083"
	log.Printf("cms example listening on http://%s", addr)
	log.Printf("open http://%s/", addr)
	log.Fatal(http.ListenAndServe(addr, newServer(renderer())))
}

func newServer(renderer *blade.Engine) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	for _, page := range pages {
		page := page
		mux.HandleFunc(page.Path, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != page.Path {
				http.NotFound(w, r)
				return
			}

			renderHTML(w, renderer, page)
		})
	}

	return mux
}

func renderHTML(w http.ResponseWriter, renderer *blade.Engine, page pageSpec) {
	output, err := renderer.RenderFile(page.View, viewContext(page))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(output))
}

func renderer() *blade.Engine {
	return blade.New(
		blade.WithFS(blade.LayeredFS{
			Layers: []blade.TemplateFS{
				blade.OSFS{
					Root: filepath.Join(exampleRoot(), "themes/default/views"),
					Namespaces: map[string]string{
						"cms": filepath.Join(exampleRoot(), "themes/default/views/cms"),
					},
				},
				blade.OSFS{
					Root: filepath.Join(exampleRoot(), "views"),
				},
				blade.OSFS{
					Root: filepath.Join(exampleRoot(), "modules/cms/views"),
					Namespaces: map[string]string{
						"cms": filepath.Join(exampleRoot(), "modules/cms/views"),
					},
				},
			},
		}),
		blade.WithDirective("badge", func(expression string) (string, error) {
			return fmt.Sprintf(`<span class="badge">{{ bladeEcho . %q }}</span>`, expression), nil
		}),
		blade.WithBlockDirective("panel", "", func(expression string, body string) (string, error) {
			return `<section class="directive-panel" data-kind="` + template.HTMLEscapeString(expression) + `">` + body + `</section>`, nil
		}),
		blade.WithDataDirective("arc", renderArticles),
		blade.WithDataDirective("site", renderSite),
		blade.WithDataDirective("arc_page", renderArticlePage),
		blade.WithDataDirective("nav", renderNav),
		blade.WithTag("pager", renderPager),
		blade.WithTag("crumb", renderCrumb),
	)
}

func viewContext(page pageSpec) ViewContext {
	return ViewContext{
		Site: SiteContext{
			Name: "Sparq CMS",
			URL:  "https://example.test",
		},
		Route: RouteContext{Path: page.Path},
		Page: PageContext{
			Title:      page.Title,
			Summary:    page.Summary,
			CategoryID: page.CategoryID,
			Pager:      page.Pager,
			Items: []any{
				blade.Item(blade.Fields().Set("title", "品牌化列表").Set("status", "推荐")),
				blade.Item(blade.Fields().Set("title", "空状态分支").Set("status", "内置")),
				blade.Item(blade.Fields().Set("title", "范围计数循环").Set("status", "轻量")),
			},
			Include: map[string]any{
				"title": "include 接收第二参数",
				"body":  "局部模板可以读取传入数据，也可以继续读取根上下文。",
			},
			Card: map[string]any{
				"title": "组件标题来自第二参数",
				"kind":  "component-demo",
			},
		},
		Crumb: []Crumb{
			{Title: "首页", URL: "/"},
			{Title: page.Title, URL: page.Path},
		},
		Docs: page.Docs,
	}
}

func renderArticles(call blade.DataTagCall) (blade.DataTagResult, error) {
	ctx := call.Context().(ViewContext)
	cid, err := call.GetIntDefault("cid", ctx.Page.CategoryID)
	if err != nil {
		return blade.DataTagResult{}, err
	}
	limit, err := call.GetIntDefault("limit", 10)
	if err != nil {
		return blade.DataTagResult{}, err
	}

	items := []any{
		blade.Item(
			blade.Fields().
				Set("id", 1).
				Set("title", fmt.Sprintf("栏目 %d 文章 A", cid)).
				Set("url", fmt.Sprintf("/category/%d/a", cid)).
				Set("summary", "文章 A 摘要").
				Set("time", "2026-06-13"),
		),
		blade.Item(
			blade.Fields().
				Set("id", 2).
				Set("title", fmt.Sprintf("栏目 %d 文章 B", cid)).
				Set("url", fmt.Sprintf("/category/%d/b", cid)).
				Set("summary", "文章 B 摘要").
				Set("time", "2026-06-12"),
		),
		blade.Item(
			blade.Fields().
				Set("id", 3).
				Set("title", fmt.Sprintf("栏目 %d 文章 C", cid)).
				Set("url", fmt.Sprintf("/category/%d/c", cid)).
				Set("summary", "文章 C 摘要").
				Set("time", "2026-06-11"),
		),
	}
	if limit < len(items) {
		items = items[:limit]
	}

	return blade.Items(items...), nil
}

func renderSite(call blade.DataTagCall) (blade.DataTagResult, error) {
	key, err := call.GetStringDefault("key", "default")
	if err != nil {
		return blade.DataTagResult{}, err
	}
	if key == "missing" {
		return blade.EmptyResult(), nil
	}

	return blade.ValueOf(
		blade.Fields().
			Set("name", "Sparq CMS").
			Set("url", "https://example.test").
			Set("icp", "沪 ICP 备 20260613 号"),
	), nil
}

func renderArticlePage(call blade.DataTagCall) (blade.DataTagResult, error) {
	cid, err := call.GetIntDefault("cid", 0)
	if err != nil {
		return blade.DataTagResult{}, err
	}
	limit, err := call.GetIntDefault("limit", 10)
	if err != nil {
		return blade.DataTagResult{}, err
	}

	items := []any{
		blade.Item(blade.Fields().Set("id", 1).Set("title", fmt.Sprintf("栏目 %d 分页文章 A", cid)).Set("url", fmt.Sprintf("/category/%d/page-a", cid))),
		blade.Item(blade.Fields().Set("id", 2).Set("title", fmt.Sprintf("栏目 %d 分页文章 B", cid)).Set("url", fmt.Sprintf("/category/%d/page-b", cid))),
		blade.Item(blade.Fields().Set("id", 3).Set("title", fmt.Sprintf("栏目 %d 分页文章 C", cid)).Set("url", fmt.Sprintf("/category/%d/page-c", cid))),
	}
	if limit < len(items) {
		items = items[:limit]
	}

	return blade.Object(
		blade.Fields().
			Set("items", items).
			Set("pager", blade.Fields().Set("page", 1).Set("size", limit).Set("total", 42)).
			Set("category", blade.Fields().Set("id", cid).Set("title", fmt.Sprintf("栏目 %d", cid))),
	), nil
}

func renderNav(call blade.DataTagCall) (blade.DataTagResult, error) {
	return blade.Items(
		blade.Item(
			blade.Fields().
				Set("title", "基础指令").
				Set("url", "/directives").
				Set("children", []any{}),
		),
		blade.Item(
			blade.Fields().
				Set("title", "模板组合").
				Set("url", "/templates").
				Set("children", []any{
					blade.Item(blade.Fields().Set("title", "局部模板").Set("url", "/templates#include")),
					blade.Item(blade.Fields().Set("title", "组件插槽").Set("url", "/templates#component")),
				}),
		),
		blade.Item(
			blade.Fields().
				Set("title", "数据处理").
				Set("url", "/data").
				Set("children", []any{}),
		),
		blade.Item(
			blade.Fields().
				Set("title", "文章列表").
				Set("url", "/articles").
				Set("children", []any{
					blade.Item(blade.Fields().Set("title", "公司动态").Set("url", "/articles/company")),
					blade.Item(blade.Fields().Set("title", "产品更新").Set("url", "/articles/product")),
				}),
		),
		blade.Item(
			blade.Fields().
				Set("title", "导航子级").
				Set("url", "/navigation").
				Set("children", []any{}),
		),
		blade.Item(
			blade.Fields().
				Set("title", "分页").
				Set("url", "/pagination").
				Set("children", []any{}),
		),
		blade.Item(
			blade.Fields().
				Set("title", "自定义扩展").
				Set("url", "/custom").
				Set("children", []any{}),
		),
	), nil
}

func renderPager(call blade.TagCall) (template.HTML, error) {
	pager, ok := call.Get("for")
	if !ok {
		ctx := call.Context().(ViewContext)
		pager = ctx.Page.Pager
	}

	view, err := call.GetStringDefault("template", "cms:partials.pager")
	if err != nil {
		return "", err
	}
	className, err := call.GetStringDefault("class", "pager")
	if err != nil {
		return "", err
	}

	return call.Render(view, map[string]any{
		"pager": pager,
		"class": className,
	})
}

func renderCrumb(call blade.TagCall) (template.HTML, error) {
	crumb, ok := call.Get("for")
	if !ok {
		ctx := call.Context().(ViewContext)
		crumb = ctx.Crumb
	}

	view, err := call.GetStringDefault("template", "partials.crumb")
	if err != nil {
		return "", err
	}

	return call.Render(view, map[string]any{
		"crumb": crumb,
	})
}

func exampleRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot resolve example root")
	}

	return filepath.Dir(file)
}
