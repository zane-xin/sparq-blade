package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	blade "github.com/zane-xin/sparq-blade"
)

type ViewContext struct {
	Site  SiteContext
	Route RouteContext
	Page  PageData
	Meta  PageMeta
}

type SiteContext struct {
	Name string
}

type RouteContext struct {
	Path string
}

type PageMeta struct {
	Title       string
	Keywords    string
	Description string
}

type PageData struct {
	Title        string
	Summary      string
	RouteKey     string
	CategoryID   int
	CategoryName string
	Items        []map[string]any
	IncludeData  map[string]any
	CardData     map[string]any
	TemplateCode string
	GoCode       string
}

type navItem struct {
	Title       string
	URL         string
	Description string
	Template    string
	Meta        PageMeta
}

var navItems = []navItem{
	{Title: "首页", URL: "/", Description: "项目应用场景入口、导航和使用流程总览。", Template: "home", Meta: PageMeta{Title: "Blade 示例操作文档", Keywords: "blade,go,模板引擎", Description: "Blade Go 模板库的中文操作文档和 HTTP 示例。"}},
	{Title: "基础渲染", URL: "/basic", Description: "创建 Engine，渲染文件模板，并使用 layout、include、sq:list。", Template: "basic", Meta: PageMeta{Title: "基础渲染", Keywords: "blade,render,layout", Description: "说明基础模板渲染、布局和循环的使用方式。"}},
	{Title: "组件与插槽", URL: "/components", Description: "使用 @component、@slot、@each 组织可复用模板片段。", Template: "component-page", Meta: PageMeta{Title: "组件与插槽", Keywords: "blade,component,directive", Description: "说明组件、插槽和自定义指令的使用方式。"}},
	{Title: "数据标签", URL: "/data-tags", Description: "通过 @sq:arc 读取业务数据，并在模板内使用 field 和 loop。", Template: "data-tags", Meta: PageMeta{Title: "数据标签", Keywords: "blade,data tag,cms", Description: "说明数据标签、参数读取和列表渲染。"}},
	{Title: "请求上下文", URL: "/request-context", Description: "让 @sq:title、@sq:seo、@sq 读取当前访问路径和页面配置。", Template: "request-context", Meta: PageMeta{Title: "请求上下文", Keywords: "blade,context,route", Description: "说明指令如何通过 call.Context() 读取请求上下文。"}},
	{Title: "语法定制", URL: "/syntax", Description: "配置 {{ }}、{!! !!}、{{-- --}} 的替代分隔符。", Template: "syntax", Meta: PageMeta{Title: "语法定制", Keywords: "blade,syntax,delimiter", Description: "说明如何通过 WithSyntax 定制模板输出、原样输出和注释分隔符。"}},
	{Title: "HTTP 服务", URL: "/http", Description: "在 net/http 中创建共享 Engine，并按请求构造 ViewContext。", Template: "http", Meta: PageMeta{Title: "HTTP 服务", Keywords: "blade,http,engine", Description: "说明 HTTP 服务中集成 Blade Engine 的方式。"}},
}

func main() {
	renderer := blade.New(
		blade.WithFS(blade.OSFS{Root: viewsRoot()}),
		blade.WithTag("title", renderTitle),
		blade.WithTag("seo", renderSEO),
		blade.WithDataDirective("nav", renderNav),
		blade.WithDataDirective("scenarios", renderScenarios),
		blade.WithDataDirective("steps", renderSteps),
		blade.WithDataDirective("arc", renderArticles),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	for _, item := range navItems {
		item := item
		mux.HandleFunc(item.URL, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != item.URL {
				http.NotFound(w, r)
				return
			}
			render(w, renderer, item, r.URL.Path)
		})
	}

	addr := "127.0.0.1:8082"
	log.Printf("docs example listening on http://%s", addr)
	log.Printf("open http://%s/", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func render(w http.ResponseWriter, renderer *blade.Engine, item navItem, path string) {
	ctx := ViewContext{
		Site:  SiteContext{Name: "Blade 示例操作文档"},
		Route: RouteContext{Path: path},
		Page:  pageData(item),
		Meta:  item.Meta,
	}

	output, err := renderer.RenderFile(item.Template, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(output))
}

func pageData(item navItem) PageData {
	data := PageData{
		RouteKey: strings.TrimPrefix(item.URL, "/"),
		Title:    item.Title,
		Summary:  item.Description,
		Items: []map[string]any{
			{"name": "Parser"},
			{"name": "Engine"},
			{"name": "Runtime"},
		},
		IncludeData: map[string]any{
			"title": "include 传入的标题",
			"body":  "局部模板可以读取传入数据，同时仍然能访问根上下文。",
		},
		CardData: map[string]any{
			"title": "component 传入的标题",
		},
	}
	if data.RouteKey == "" {
		data.RouteKey = "home"
	}

	switch item.Template {
	case "basic":
		data.TemplateCode = `@extends('layout')

@section('content')
  @include('partials.doc-note', Page.IncludeData)
  @list(Page.Items, as=item)
    <li>{{ item.name }}</li>
  @empty
    <li>暂无内容</li>
  @end
@endsection`
		data.GoCode = `renderer := blade.New(blade.WithFS(blade.OSFS{Root: "./views"}))
output, err := renderer.RenderFile("basic", viewContext)`
	case "component-page":
		data.TemplateCode = `@component('components.doc-card', Page.CardData)
  组件正文
@endcomponent`
		data.GoCode = `renderer := blade.New(
    blade.WithDirective("badge", badgeCompiler),
    blade.WithBlockDirective("panel", "", panelCompiler),
)`
	case "data-tags":
		data.CategoryID = 3
		data.CategoryName = "新闻"
		data.TemplateCode = `@sq:arc(cid=3, limit=2, as=article)
  <li>{{ article.title }}</li>
@end`
		data.GoCode = `blade.WithDataDirective("arc", func(call blade.DataTagCall) (blade.DataTagResult, error) {
    cid, _ := call.GetInt("cid")
    return articleList(cid), nil
})`
	case "request-context":
		data.CategoryID = 7
		data.CategoryName = "产品"
		data.TemplateCode = `@sq:arc(limit=2, as=article)
  <li>{{ article.title }}</li>
@end`
		data.GoCode = `cid, _ := call.GetIntDefault("cid", 0)
if cid == 0 {
    ctx := call.Context().(ViewContext)
    cid = ctx.Page.CategoryID
}`
	case "syntax":
		data.TemplateCode = `[[ title ]]
[[! html !]]
[[-- 这段不会输出 --]]`
		data.GoCode = `renderer := blade.New(
    blade.WithSyntax(blade.Syntax{
        EchoOpen:     "[[",
        EchoClose:    "]]",
        RawOpen:      "[[!",
        RawClose:     "!]]",
        CommentOpen:  "[[--",
        CommentClose: "--]]",
    }),
)`
	case "http":
		data.TemplateCode = `mux.HandleFunc("/news", func(w http.ResponseWriter, r *http.Request) {
    output, _ := renderer.RenderFile("page", requestContext(r))
    w.Write([]byte(output))
})`
		data.GoCode = `var renderer = blade.New(blade.WithFS(blade.OSFS{Root: "./views"}))
// 每次请求只创建 ViewContext；Engine 可以复用。`
	}

	return data
}

func renderTitle(call blade.TagCall) (template.HTML, error) {
	ctx, err := viewContext(call.Context())
	if err != nil {
		return "", err
	}
	return template.HTML(template.HTMLEscapeString(ctx.Meta.Title)), nil
}

func renderSEO(call blade.TagCall) (template.HTML, error) {
	ctx, err := viewContext(call.Context())
	if err != nil {
		return "", err
	}

	return template.HTML(fmt.Sprintf(
		`<meta name="keywords" content="%s"><meta name="description" content="%s">`,
		template.HTMLEscapeString(ctx.Meta.Keywords),
		template.HTMLEscapeString(ctx.Meta.Description),
	)), nil
}

func renderNav(call blade.DataTagCall) (blade.DataTagResult, error) {
	ctx, err := viewContext(call.Context())
	if err != nil {
		return blade.DataTagResult{}, err
	}

	items := make([]any, 0, len(navItems))
	for _, item := range navItems {
		items = append(items, blade.Item(blade.Fields().
			Set("title", item.Title).
			Set("url", item.URL).
			Set("description", item.Description).
			Set("active", ctx.Route.Path == item.URL),
		))
	}

	return blade.Items(items...), nil
}

func renderScenarios(call blade.DataTagCall) (blade.DataTagResult, error) {
	items := make([]any, 0, len(navItems)-1)
	for _, item := range navItems[1:] {
		items = append(items, blade.Item(blade.Fields().
			Set("title", item.Title).
			Set("url", item.URL).
			Set("description", item.Description),
		))
	}

	return blade.Items(items...), nil
}

func renderSteps(call blade.DataTagCall) (blade.DataTagResult, error) {
	page, err := call.GetStringDefault("page", "basic")
	if err != nil {
		return blade.DataTagResult{}, err
	}

	steps := map[string][]string{
		"basic":           {"创建 Engine", "配置模板文件系统", "准备 ViewContext", "调用 RenderFile 输出 HTML"},
		"components":      {"创建组件模板", "使用 @component 渲染组件", "通过第二参数传入组件数据", "按需要通过 slot 传入正文或具名内容"},
		"data-tags":       {"注册 WithDataDirective", "从参数读取 cid/limit", "返回 Items 或 ValueOf", "在模板中用 @sq:arc 渲染数据"},
		"request-context": {"定义请求上下文结构", "控制器写入 ViewContext", "指令通过 call.Context() 读取上下文", "显式参数优先于上下文推导"},
		"syntax":          {"在 Engine 上设置 WithSyntax", "保留默认指令语法 @if/@sq", "替换输出、原样输出和注释分隔符", "按项目规范统一模板写法"},
		"http":            {"启动 HTTP 服务", "复用同一个 Engine", "每次请求构造 ViewContext", "RenderFile 输出响应 HTML"},
	}

	values := steps[page]
	items := make([]any, 0, len(values))
	for index, step := range values {
		items = append(items, blade.Item(blade.Fields().
			Set("index", index+1).
			Set("title", step),
		))
	}

	return blade.Items(items...), nil
}

func renderArticles(call blade.DataTagCall) (blade.DataTagResult, error) {
	ctx, err := viewContext(call.Context())
	if err != nil {
		return blade.DataTagResult{}, err
	}

	cid, err := call.GetIntDefault("cid", ctx.Page.CategoryID)
	if err != nil {
		return blade.DataTagResult{}, err
	}
	if cid == 0 {
		cid = 1
	}

	limit, err := call.GetIntDefault("limit", 3)
	if err != nil {
		return blade.DataTagResult{}, err
	}

	items := []any{
		blade.Item(blade.Fields().Set("title", fmt.Sprintf("分类 %d 示例文章 A", cid)).Set("url", fmt.Sprintf("/docs/%d/a", cid))),
		blade.Item(blade.Fields().Set("title", fmt.Sprintf("分类 %d 示例文章 B", cid)).Set("url", fmt.Sprintf("/docs/%d/b", cid))),
		blade.Item(blade.Fields().Set("title", fmt.Sprintf("分类 %d 示例文章 C", cid)).Set("url", fmt.Sprintf("/docs/%d/c", cid))),
	}
	if limit < len(items) {
		items = items[:limit]
	}

	return blade.Items(items...), nil
}

func viewContext(value any) (ViewContext, error) {
	ctx, ok := value.(ViewContext)
	if ok {
		return ctx, nil
	}

	ptr, ok := value.(*ViewContext)
	if ok && ptr != nil {
		return *ptr, nil
	}

	return ViewContext{}, fmt.Errorf("unexpected view context: %T", value)
}

func viewsRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "views")
}
