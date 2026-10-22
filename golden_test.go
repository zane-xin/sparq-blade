package blade_test

import (
	"fmt"
	"testing"

	blade "github.com/zane-xin/sparq-blade"
)

func TestRenderStringGolden(t *testing.T) {
	renderer := blade.New(
		blade.WithDirective("badge", func(expression string) (string, error) {
			return fmt.Sprintf(`<strong>{{ bladeEcho . %q }}</strong>`, expression), nil
		}),
		blade.WithBlockDirective("panel", "", func(_ string, body string) (string, error) {
			return `<section>` + body + `</section>`, nil
		}),
		blade.WithDataDirective("arc", func(call blade.DataTagCall) (blade.DataTagResult, error) {
			limit, err := call.GetIntDefault("limit", 2)
			if err != nil {
				return blade.DataTagResult{}, err
			}

			items := make([]any, 0, limit)
			for index := 0; index < limit; index++ {
				items = append(items, blade.Item(blade.Fields().Set("title", fmt.Sprintf("item-%d", index+1))))
			}

			return blade.Items(items...), nil
		}),
	)

	cases := []struct {
		name   string
		source string
		data   any
		want   string
	}{
		{
			name:   "echo and raw",
			source: `<h1>{{ title }}</h1><div>{!! html !!}</div>`,
			data: map[string]any{
				"title": "<Blade>",
				"html":  "<b>safe</b>",
			},
			want: `<h1>&lt;Blade&gt;</h1><div><b>safe</b></div>`,
		},
		{
			name:   "echo without dollar prefix",
			source: `<h1>{{ title }}</h1><p>{{ user.name }}</p>`,
			data: map[string]any{
				"title": "Blade",
				"user":  map[string]any{"name": "Alice"},
			},
			want: `<h1>Blade</h1><p>Alice</p>`,
		},
		{
			name:   "conditionals",
			source: `@if(count > 2)many@elseif(count == 2)pair@else single@endif`,
			data:   map[string]any{"count": 2},
			want:   `pair`,
		},
		{
			name:   "conditionals without dollar prefix",
			source: `@if(visible && count == 2)shown@else hidden@endif`,
			data:   map[string]any{"visible": true, "count": 2},
			want:   `shown`,
		},
		{
			name:   "foreach loop state",
			source: `@foreach(items as item)<span>{{ loop.index }}:{{ loop.first }}:{{ loop.last }}:{{ item }}</span>@endforeach`,
			data:   map[string]any{"items": []string{"A", "B"}},
			want:   `<span>0:true:false:A</span><span>1:false:true:B</span>`,
		},
		{
			name:   "foreach without dollar prefix",
			source: `@foreach(items as item)<span>{{ loop.index }}:{{ item.name }}</span>@endforeach`,
			data: map[string]any{
				"items": []map[string]any{
					{"name": "A"},
					{"name": "B"},
				},
			},
			want: `<span>0:A</span><span>1:B</span>`,
		},
		{
			name:   "for loop",
			source: `@for(i in 1..3){{ i }}@endfor`,
			want:   `123`,
		},
		{
			name:   "list empty",
			source: `@list(items, as=item)<span>{{ item }}</span>@empty empty@endlist`,
			data:   map[string]any{"items": []string{}},
			want:   `empty`,
		},
		{
			name:   "children with wrap",
			source: `@list(items, as=item){{ item.title }}@children(as=child, wrap="ul", class="children")<li>{{ child.title }}</li>@endchildren@endlist`,
			data: map[string]any{
				"items": []map[string]any{
					{
						"title": "A",
						"children": []map[string]any{
							{"title": "A1"},
						},
					},
				},
			},
			want: `A<ul class="children"><li>A1</li></ul>`,
		},
		{
			name:   "custom directives",
			source: `@sq:panel()@sq:badge(name)@end`,
			data:   map[string]any{"name": "Blade"},
			want:   `<section><strong>Blade</strong></section>`,
		},
		{
			name:   "data tag block",
			source: `<ul>@sq:arc(limit=2)<li>{{ loop.iteration }}:{{ field.title }}</li>@end</ul>`,
			want:   `<ul><li>1:item-1</li><li>2:item-2</li></ul>`,
		},
		{
			name:   "data tag block without dollar prefix",
			source: `<ul>@sq:arc(limit=2, as=field)<li>{{ loop.iteration }}:{{ field.title }}</li>@end</ul>`,
			want:   `<ul><li>1:item-1</li><li>2:item-2</li></ul>`,
		},
		{
			name:   "data tag out",
			source: `@sq:arc(limit=1, out='articles')@list(articles, as=item)<b>{{ item.title }}</b>@end`,
			want:   `<b>item-1</b>`,
		},
		{
			name:   "data tag out without dollar prefix",
			source: `@sq:arc(limit=1, out=articles)@list(articles, as=item)<b>{{ item.title }}</b>@end`,
			want:   `<b>item-1</b>`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderer.RenderString(tt.source, tt.data)
			if err != nil {
				t.Fatalf("RenderString() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected output:\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
