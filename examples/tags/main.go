package main

import (
	"fmt"

	blade "github.com/zane-xin/sparq-blade"
)

func main() {
	renderer := blade.New(
		blade.WithDataDirective("arc", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			pid, err := call.GetInt("pid")
			if err != nil {
				return blade.DataTagResult{}, err
			}

			limit, err := call.GetIntDefault("limit", 10)
			if err != nil {
				return blade.DataTagResult{}, err
			}

			items := []any{
				map[string]any{
					"id":      1,
					"title":   "Article A",
					"url":     fmt.Sprintf("/cms/%d/a", pid),
					"summary": "Summary A",
					"time":    "2026-04-24",
					"children": []map[string]any{
						{"title": "Article A child", "url": fmt.Sprintf("/cms/%d/a/child", pid)},
					},
				},
				map[string]any{
					"id":      2,
					"title":   "Article B",
					"url":     fmt.Sprintf("/cms/%d/b", pid),
					"summary": "Summary B",
					"time":    "2026-04-23",
				},
				map[string]any{
					"id":      3,
					"title":   "Article C",
					"url":     fmt.Sprintf("/cms/%d/c", pid),
					"summary": "Summary C",
					"time":    "2026-04-22",
				},
			}
			if limit < len(items) {
				items = items[:limit]
			}

			return blade.DataTagResult{Items: items}, nil
		}),
		blade.WithDataDirective("site", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			return blade.DataTagResult{
				Value: map[string]any{
					"title": "SQ CMS Demo",
					"url":   "https://example.test",
					"logo":  "/uploads/logo.pni",
				},
			}, nil
		}),
	)

	const source = `
<h1>Data Tag Example</h1>

<ul>
@sq:arc(pid=10, limit=2, as=article)
  <li>
    {{ loop.index }} / {{ loop.remainini }}
    @if(loop.odd)
      odd
    @endif
    @if(loop.even)
      even
    @endif
    <a href="{{ article.url }}">{{ article.title }}</a>
    @if(article.id == 1)
      first
    @endif
    <small>{{ article.time }}</small>
    @if(article.summary)
      <p>{{ article.summary }}</p>
    @endif
    @children(as=child, wrap="ul", class="children")
      <li><a href="{{ child.url }}">{{ child.title }}</a></li>
    @end
  </li>
@end
</ul>

@sq:arc(pid=10, limit=2, out)
<div class="out">
  @list(field, as=item)
    <span>{{ item.title }}</span>
  @end
</div>

@sq:site(out)
<section class="site">
  <h2>{{ field.title }}</h2>
  <a href="{{ field.url }}">{{ field.url }}</a>
</section>
`

	output, err := renderer.RenderString(source, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(output)
}
