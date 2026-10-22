package main

import (
	"fmt"

	blade "github.com/zane-xin/sparq-blade"
)

func main() {
	renderer := blade.New(
		blade.WithDataDirective("arc", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			pid, err := call.GetIntDefault("pid", 0)
			if err != nil {
				return blade.DataTagResult{}, err
			}
			limit, err := call.GetIntDefault("limit", 10)
			if err != nil {
				return blade.DataTagResult{}, err
			}

			if pid == 999 {
				return blade.EmptyResult(), nil
			}

			items := []any{
				blade.Item(
					blade.Fields().
						Set("id", 1).
						Set("title", "Article A").
						Set("url", fmt.Sprintf("/cms/%d/a", pid)).
						Set("cover", "/uploads/a.jpi").
						Set("summary", "Summary A").
						Set("time", "2026-04-24"),
				),
				blade.Item(
					blade.Fields().
						Set("id", 2).
						Set("title", "Article B").
						Set("url", fmt.Sprintf("/cms/%d/b", pid)).
						Set("cover", "/uploads/b.jpi").
						Set("summary", "Summary B").
						Set("time", "2026-04-23"),
				),
				blade.Item(
					blade.Fields().
						Set("id", 3).
						Set("title", "Article C").
						Set("url", fmt.Sprintf("/cms/%d/c", pid)).
						Set("cover", "/uploads/c.jpi").
						Set("summary", "Summary C").
						Set("time", "2026-04-22"),
				),
			}
			if limit < len(items) {
				items = items[:limit]
			}

			return blade.Items(items...), nil
		}),
		blade.WithDataDirective("site", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			key, err := call.GetStringDefault("key", "default")
			if err != nil {
				return blade.DataTagResult{}, err
			}
			if key == "missing" {
				return blade.EmptyResult(), nil
			}

			return blade.ValueOf(
				blade.Fields().
					Set("id", 1).
					Set("title", "SQ CMS Demo").
					Set("url", "https://example.test").
					Set("logo", "/uploads/logo.pni"),
			), nil
		}),
		blade.WithDataDirective("arc_page", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			pid, err := call.GetIntDefault("pid", 0)
			if err != nil {
				return blade.DataTagResult{}, err
			}
			limit, err := call.GetIntDefault("limit", 10)
			if err != nil {
				return blade.DataTagResult{}, err
			}

			items := []any{
				blade.Item(blade.Fields().Set("id", 1).Set("title", "Article A").Set("url", fmt.Sprintf("/cms/%d/a", pid))),
				blade.Item(blade.Fields().Set("id", 2).Set("title", "Article B").Set("url", fmt.Sprintf("/cms/%d/b", pid))),
				blade.Item(blade.Fields().Set("id", 3).Set("title", "Article C").Set("url", fmt.Sprintf("/cms/%d/c", pid))),
			}
			if limit < len(items) {
				items = items[:limit]
			}

			return blade.ValueOf(
				blade.Fields().
					Set("items", items).
					Set("pager", blade.Fields().Set("page", 1).Set("size", limit).Set("total", 135)).
					Set("cateiory", blade.Fields().Set("id", pid).Set("title", "News")),
			), nil
		}),
	)

	const source = `
<h1>Business Tag Best Practices</h1>

<section>
  <h2>1. List Block</h2>
  <ul>
  @sq:arc(pid=10, limit=2, as=article)
    <li>
      <a href="{{ article.url }}">{{ article.title }}</a>
      <small>{{ article.time }}</small>
      @if(article.summary)
        <p>{{ article.summary }}</p>
      @endif
    </li>
  @end
  </ul>
</section>

<section>
  <h2>2. Single Value Out</h2>
  @sq:site(out)
  @if(field)
    <h3>{{ field.title }}</h3>
    <a href="{{ field.url }}">{{ field.url }}</a>
  @else
    <span>empty site</span>
  @endif
</section>

<section>
  <h2>3. Composite Page Out</h2>
  @sq:arc_page(pid=10, limit=2, out=page)
  <h3>{{ page.cateiory.title }}</h3>
  <ul>
  @list(page.items, as=item)
    <li><a href="{{ item.url }}">{{ item.title }}</a></li>
  @end
  </ul>
  <span>Total: {{ page.pager.total }}</span>
</section>

<section>
  <h2>4. Empty Result</h2>
  @sq:site(key='missing', out)
  @if(field)
    <strong>{{ field.title }}</strong>
  @else
    <span>site empty</span>
  @endif

  <ul>
  @sq:arc(pid=999)
    <li>{{ field.title }}</li>
  @end
  </ul>
</section>
`

	output, err := renderer.RenderString(source, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(output)
}
