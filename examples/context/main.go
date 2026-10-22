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
	Meta  PageMeta
	Data  map[string]any
}

type SiteContext struct {
	Name string
}

type RouteContext struct {
	Path  string
	Alias string
}

type PageContext struct {
	Category *Category
}

type Category struct {
	ID    int
	Alias string
	Name  string
}

type PageMeta struct {
	Title       string
	Keywords    string
	Description string
}

func main() {
	renderer := blade.New(
		blade.WithFS(blade.OSFS{Root: viewsRoot()}),
		blade.WithTag("title", renderTitle),
		blade.WithTag("seo", renderSEO),
		blade.WithDataDirective("arc", renderArticles),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/news", renderPage(renderer, Category{
		ID:    7,
		Alias: "news",
		Name:  "新闻",
	}, PageMeta{
		Title:       "新闻列表",
		Keywords:    "新闻,文章,动态",
		Description: "新闻栏目通过当前访问上下文自动推导文章列表。",
	}))
	mux.HandleFunc("/products", renderPage(renderer, Category{
		ID:    12,
		Alias: "products",
		Name:  "产品",
	}, PageMeta{
		Title:       "产品列表",
		Keywords:    "产品,方案,案例",
		Description: "产品栏目使用同一套模板和不同的请求上下文。",
	}))

	addr := "127.0.0.1:8081"
	log.Printf("context example listening on http://%s", addr)
	log.Printf("open http://%s/news", addr)
	log.Printf("open http://%s/products", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func renderPage(renderer *blade.Engine, category Category, meta PageMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := ViewContext{
			Site: SiteContext{Name: "SQ Admin"},
			Route: RouteContext{
				Path:  r.URL.Path,
				Alias: category.Alias,
			},
			Page: PageContext{
				Category: &category,
			},
			Meta: meta,
		}

		output, err := renderer.RenderFile("page", ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(output))
	}
}

func renderTitle(call blade.TagCall) (template.HTML, error) {
	ctx, err := viewContext(call.Context())
	if err != nil {
		return "", err
	}

	title := ctx.Meta.Title
	if title == "" && ctx.Page.Category != nil {
		title = ctx.Page.Category.Name
	}
	if title == "" {
		title = ctx.Site.Name
	}

	return template.HTML(template.HTMLEscapeString(title)), nil
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

func renderArticles(call blade.DataTagCall) (blade.DataTagResult, error) {
	ctx, err := viewContext(call.Context())
	if err != nil {
		return blade.DataTagResult{}, err
	}

	cid, err := call.GetIntDefault("cid", 0)
	if err != nil {
		return blade.DataTagResult{}, err
	}
	if cid == 0 {
		if ctx.Page.Category == nil {
			return blade.EmptyResult(), fmt.Errorf("@arc missing cid and current category")
		}
		cid = ctx.Page.Category.ID
	}

	limit, err := call.GetIntDefault("limit", 10)
	if err != nil {
		return blade.DataTagResult{}, err
	}

	items := []any{
		blade.Item(blade.Fields().Set("id", 1).Set("title", fmt.Sprintf("栏目 %d 文章 A", cid)).Set("url", fmt.Sprintf("/articles/%d/a", cid))),
		blade.Item(blade.Fields().Set("id", 2).Set("title", fmt.Sprintf("栏目 %d 文章 B", cid)).Set("url", fmt.Sprintf("/articles/%d/b", cid))),
		blade.Item(blade.Fields().Set("id", 3).Set("title", fmt.Sprintf("栏目 %d 文章 C", cid)).Set("url", fmt.Sprintf("/articles/%d/c", cid))),
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
