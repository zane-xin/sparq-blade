# Quick Start

这份文档面向第一次接入 `blade` 的使用者，目标是用最短路径跑通：

1. 渲染模板字符串
2. 渲染模板文件
3. 使用组件与插槽
4. 注册自定义指令
5. 在 HTTP 服务中接入
6. 注册业务标签

## 1. 安装

当前仓库模块名：

```go
github.com/zane-xin/sparq-blade
```

如需显式拉取依赖，可执行：

```bash
go get github.com/zane-xin/sparq-blade@latest
```

在你的项目里引入：

```go
import blade "github.com/zane-xin/sparq-blade"
```

常用场景和高级配置都只需要引入根包。底层解析、渲染和运行时实现位于 `internal` 包内，不作为稳定公开 API。

## 2. 最小字符串渲染

这是最短可运行示例：

```go
output, err := blade.RenderString(`<h1>{{ title }}</h1>`, map[string]any{
    "title": "Blade",
})
if err != nil {
    panic(err)
}

fmt.Println(output)
```

输出：

```html
<h1>Blade</h1>
```

## 3. 文件模板渲染

### 目录示例

```text
views/
  layout.tpl
  home.tpl
  footer.tpl
```

### 模板内容

`layout.tpl`

```blade
<html>
  <body>
    @yield('content')
    @include('footer')
  </body>
</html>
```

`home.tpl`

```blade
@extends('layout')

@section('content')
<h1>{{ title }}</h1>
<ul>
  @list(items, as=item)
  <li>{{ item.name }}</li>
  @empty
  <li>暂无内容</li>
  @end
</ul>
@endsection
```

`footer.tpl`

```blade
<footer>{{ brand }}</footer>
```

也可以给 include 显式传入数据：

```blade
@include('profile', user)
```

`profile.tpl` 中可以直接读取 `name` 这类由 `user` 展开的字段。

### Go 代码

```go
renderer := blade.New(
    blade.WithFS(blade.OSFS{Root: "./views"}),
)

output, err := renderer.RenderFile("home", map[string]any{
    "title": "Home",
    "items": []map[string]any{
        {"name": "Parser"},
        {"name": "Engine"},
    },
    "brand": "sparq/blade",
})
if err != nil {
    panic(err)
}
```

说明：

- 变量路径直接写 `title` 或 `article.title`；`$title` 属于不支持的旧语法
- 文件名支持不带后缀调用，例如 `RenderFile("home", ...)`
- 点号路径会映射成目录，例如 `default.view` -> `default/view`
- 如需按模块拆分视图，可使用命名空间写法，例如 `cms:default.view`
- 默认支持 `.tpl`、`.html`

命名空间目录示例：

```go
renderer := blade.New(
    blade.WithFS(blade.OSFS{
        Root: "./views",
        Namespaces: map[string]string{
            "cms": "./modules/cms/views",
        },
    }),
)
```

需要主题覆盖或应用覆盖时，可以用 `LayeredFS` 按顺序查找模板：

```go
renderer := blade.New(
    blade.WithFS(blade.LayeredFS{
        Layers: []blade.TemplateFS{
            blade.OSFS{Root: "./themes/default/views"},
            blade.OSFS{Root: "./views"},
            blade.OSFS{
                Root: "./vendor/views",
                Namespaces: map[string]string{
                    "cms": "./modules/cms/views",
                },
            },
        },
    }),
)
```

模板里仍然写稳定路径：

```blade
@include('partials.nav')
@include('cms:partials.pager')
```

`LayeredFS` 只在当前层返回未找到时继续查找下一层；`OSFS` 会拒绝 `../`、绝对路径和符号链接逃逸，模板不会越过配置的安全目录。

## 4. 组件与插槽

组件模板：

```blade
<article>
  <header>{{ title }}</header>
  <section>{{ slot }}</section>
</article>
```

调用模板：

```blade
@component('card')
  body content

  @slot('title')
    Card Title
  @endslot
@endcomponent
```

组件也可以显式传入数据：

```blade
@component('card', card)
  body content
@endcomponent
```

说明：

- 默认插槽在组件模板中通过 `slot` 使用
- 命名插槽通过 `title` / `title`、`footer` / `footer` 这类变量读取
- 第二个参数会合并到组件模板作用域中
- 插槽内容仍按调用处作用域渲染

## 5. 自定义业务指令

通过 `WithDirective` 和 `WithBlockDirective` 注册的指令同样属于业务指令。Go 侧只传局部名，模板侧固定使用 `@sq:`，不提供裸指令或自定义品牌入口。

### 单指令

```go
renderer := blade.New(
    blade.WithDirective("badge", func(expression string) (string, error) {
        return fmt.Sprintf(`<strong>{{ bladeEcho . %q }}</strong>`, expression), nil
    }),
)
```

模板里：

```blade
@sq:badge(name)
```

### 块指令

```go
renderer := blade.New(
    blade.WithBlockDirective("panel", "", func(expression string, body string) (string, error) {
        return `<section class="panel">` + body + `</section>`, nil
    }),
)
```

模板里：

```blade
@sq:panel()
  Hello {{ name }}
@end
```

## 6. HTTP 接入

最小接入方式：

```go
renderer := blade.New(
    blade.WithFS(blade.OSFS{Root: "./views"}),
)

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    output, err := renderer.RenderFile("home", map[string]any{
        "title": "Blade HTTP Demo",
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    _, _ = w.Write([]byte(output))
})
```

更完整的 HTTP 示例见：

- [examples/http/main.go](examples/http/main.go)

启动后可快速验证：

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/
curl http://127.0.0.1:8080/users
```

## 7. 业务标签

如果你需要类似 CMS 标签、插件标签这类能力，可以使用实验性业务标签 API。

取数据并在模板中循环时，使用 `WithDataDirective`：

```go
renderer := blade.New(
    blade.WithDataDirective("arc", func(call blade.DataTagCall) (blade.DataTagResult, error) {
        limit, err := call.GetIntDefault("limit", 10)
        if err != nil {
            return blade.DataTagResult{}, err
        }

        _ = limit

        return blade.Rows(
            map[string]any{"id": 1, "title": "A", "url": "/a", "summary": "Summary A"},
            map[string]any{"id": 2, "title": "B", "url": "/b", "summary": "Summary B"},
        ), nil
    }),
    blade.WithDataDirective("site", func(call blade.DataTagCall) (blade.DataTagResult, error) {
        return blade.Object(map[string]any{
            "title": "SQ CMS",
            "url":   "https://example.test",
            "logo":  "/uploads/logo.pni",
        }), nil
    }),
)
```

模板里使用品牌化数据指令：

```blade


@sq:arc(pid=10, limit=20, as=article)
  <li>
    <a href="{{ article.url }}">{{ article.title }}</a>
    @if(article.summary)
      <p>{{ article.summary }}</p>
    @endif
  </li>
@empty
  <li>暂无内容</li>
@end

@sq:arc(pid=10, limit=20, out)

@sq:arc(pid=10, limit=20, out='articleList')

@sq:site(out)
<h1>{{ field.title }}</h1>
```

直接输出 HTML 时，使用 `WithTag`：

```go
renderer := blade.New(
    blade.WithTag("pager", func(call blade.TagCall) (template.HTML, error) {
        pager, ok := call.Get("for")
        if !ok {
            ctx := call.Context().(ViewContext)
            pager = ctx.Page.Pager
        }

        className, err := call.GetStringDefault("class", "pager")
        if err != nil {
            return "", err
        }

        return renderPager(pager, className), nil
    }),
)
```

模板里：

```blade
@sq:pager()
@sq:pager(class="pager")
```

说明：

- Go 侧按目标分别注册，例如 `blade.WithDataDirective("arc", handler)`
- 模板侧的业务指令固定使用品牌入口，例如 `@sq:arc`；`@if`、`@list`、`@empty`、`@end` 等流程指令不使用 `sq:` 前缀
- 直接输出 HTML 的业务标签用 `blade.WithTag("pager", handler)`，模板侧写 `@sq:pager(...)`
- 所有外部注册指令都使用业务局部名并映射到固定的 `sq:` 入口；显式传入任何命名空间都会在初始化时失败
- 除 `out` 可单独写成裸参数外，其余参数当前使用 `key=value` 命名参数形式
- `as=article` 控制块模式下当前项变量名，默认是 `field`
- `out` 默认导出到 `field`
- `out='articleList'` 用于显式指定导出变量名
- 单指令 `out` 既可以导出列表结果，也可以导出单对象结果
- 块模式当前只接受 `Items` 列表结果，不接受单对象 `Value`
- 数据业务块会直接遍历 `Items`，块内不需要再嵌套 `@list`；空结果可以通过 `@empty` 渲染占位内容
- 使用 `out=articles` 导出数据后，才通过 `@list(articles, as=article)` 等流程指令继续处理
- 块模式当前会注入 `loop.index`、`loop.iteration`、`loop.count`、`loop.remainini`、`loop.first`、`loop.last`、`loop.odd`、`loop.even`
- 如需减少 handler 中重复的 `DataTagResult{...}`，可使用 `blade.Rows()`、`blade.Object()`、`blade.EmptyResult()`
- `Rows` 支持 map、struct，以及 map/struct slice；需要链式构造字段时仍可使用 `blade.Fields()`、`blade.Item()`、`blade.Items()`、`blade.ValueOf()`
- 推荐业务字段先尽量统一，例如列表项使用 `id/title/url/cover/summary/time`，单对象使用 `id/title/url/logo`
- 分页或复合数据推荐走 `out=page` 这类显式变量，返回 `Value` 对象，再访问 `page.items`、`page.pager`
- 单对象标签查不到数据时，推荐直接返回 `DataTagResult{}`，模板侧可用 `@if(field)` 判空
- 列表标签块模式查不到数据时，直接返回空 `Items` 即可，模板自然无输出
- 数据业务块使用统一流程结束指令 `@end`
- 该能力当前仍属于实验性 API

## 8. 推荐起步顺序

如果你是第一次接这个库，建议顺序是：

1. 先跑 `go run ./examples/basic`
2. 再跑 `go run ./examples/components`
3. 再跑 `go run ./examples/tags`
4. 再跑 `go run ./examples/tags-best-practices`
5. 如果是 Web 项目，再跑 `go run ./examples/http`

## 9. 常见坑

### 1. 模板循环优先使用 `@list`

普通列表渲染优先使用 `@list(items, as=item)` 和 `@empty`。`@foreach` 保留给高阶模板场景，计数循环使用 `@for(i in 1..5)`，当前不支持 `@while`。

### 2. 组件当前是视图组件，不是完整属性系统

当前重点是模板复用，不是完整的组件参数协议。

### 3. 自定义指令先保持简单

优先实现“包装型”指令，不建议一开始就设计复杂 DSL。

## 10. CMS 接入前建议

如果你准备把当前业务标签能力接到真实 CMS 项目里，建议在正式迁移前先过一遍：

- [docs/integration-checklist.md](docs/integration-checklist.md)

重点看：

- 是否接受 `field` 作为块模式和裸 `out` 的默认变量
- 是否接受列表用 `Items`、单对象用 `Value`
- 是否接受分页等复合数据统一走 `out=page`
- 空结果时模板层的判空写法是否自然

如需后续集中查阅，可直接看：

- [docs/README.md](docs/README.md)

### 业务标签当前还是实验性能力

推荐先用于明确受控的业务标签场景，例如 CMS 标签。

## 11. 下一步文档

如果你已经跑通 Quick Start，可以继续看：

- [README.md](README.md)
- [examples/README.md](examples/README.md)
- [docs/support.md](docs/support.md)
