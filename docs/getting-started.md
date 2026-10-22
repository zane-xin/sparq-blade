# Getting Started

这份文档面向第一次接入 `blade` 的使用者，目标是用最短路径跑通当前能力。

## 1. 安装

模块名：

```go
github.com/zane-xin/sparq-blade
```

安装：

```bash
go get github.com/zane-xin/sparq-blade@latest
```

引入：

```go
import blade "github.com/zane-xin/sparq-blade"
```

## 2. 最小字符串渲染

```go
output, err := blade.RenderString(`<h1>{{ title }}</h1>`, map[string]any{
    "title": "Blade",
})
if err != nil {
    panic(err)
}

fmt.Println(output)
```

## 3. 文件模板渲染

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
})
if err != nil {
    panic(err)
}
```

模板名规则：

- `home` 会自动匹配模板后缀
- `default.view` 会映射成目录路径
- `cms:default.view` 支持命名空间目录

## 4. 模板组合与数据传递

当前支持：

- `@include`
- `@include('view', data)`
- `@extends`
- `@section`
- `@yield`
- `@component`
- `@component('view', data)`
- `@slot`
- `@each`

常用写法：

```blade
@extends('layout')

@section('content')
  @include('partials.note', Page.IncludeData)

  @component('components.card', Page.CardData)
    组件正文
  @endcomponent
@endsection
```

数据规则：

- `@include('view')` 直接沿用当前作用域。
- `@include('view', data)` 会把 `data` 展开为局部变量，同时保留根上下文和当前循环变量。
- `@component('view', data)` 会把 `data` 展开为组件局部变量，同时保留默认 slot、具名 slot 和根上下文。
- `@extends` 只负责布局继承，不传递第二参数；布局需要的数据应放在 `RenderFile` 的根 `ViewContext` 中。

## 5. 自定义模板语法

默认语法保持 Blade 风格：

```blade
{{ title }}
{{ title }}
{!! html !!}
{!! html !!}
{{-- comment --}}
```

普通数据使用 `{{ }}` 输出，由引擎根据 HTML 上下文自动转义。`{!! !!}` 只用于应用生成或经过可靠 HTML 清洗的可信内容，不得直接输出请求参数、数据库富文本或其他外部输入。

变量路径直接写 `title` 或 `article.title`。`$title` 不再兼容，迁移已有 Blade 模板时必须移除变量前缀。

如果需要避开前端框架的 `{{ }}` 冲突，可以在创建 `Engine` 时替换输出、原样输出和注释分隔符：

```go
renderer := blade.New(
    blade.WithSyntax(blade.Syntax{
        EchoOpen:     "[[",
        EchoClose:    "]]",
        RawOpen:      "[[!",
        RawClose:     "!]]",
        CommentOpen:  "[[--",
        CommentClose: "--]]",
    }),
)
```

对应模板写法：

```blade
[[ title ]]
[[! html !]]
[[-- comment --]]
```

指令前缀仍然固定为 `@`。业务指令使用固定品牌命名空间，例如 `@sq:arc`；`@if`、`@list`、`@empty`、`@end` 等流程指令不使用品牌前缀。

## 6. 自定义业务指令

当前支持：

- `blade.WithDirective`
- `blade.WithBlockDirective`

这两类扩展也属于业务指令：Go 侧只注册 `badge`、`panel` 这类局部名，模板侧固定使用 `@sq:badge(...)`、`@sq:panel(...) ... @end`。调用方不能提供其他品牌命名空间。

适合做简单包装型语义，不建议一开始引入复杂 DSL。

## 7. 业务标签

当前推荐使用品牌化业务指令：

```blade
@sq:arc(pid=10, limit=20, as=article)
  <li>{{ article.title }}</li>
@end

@sq:site(out)
@if(field)
  <h1>{{ field.title }}</h1>
@endif
```

Go 侧按目标注册：

```go
renderer := blade.New(
    blade.WithDataDirective("arc", func(call blade.DataTagCall) (blade.DataTagResult, error) {
        return blade.Rows(
            map[string]any{"id": 1, "title": "A", "url": "/a"},
        ), nil
    }),
)
```

关键约定：

- 块模式默认当前项变量是 `field` / `field`
- 裸 `out` 默认也导出到 `field` / `field`
- 需要多个结果并存时显式写 `out=arc`、`out=page`
- 列表走 `Items`
- 单对象或复合数据走 `Value` + `out`

## 8. 请求上下文

业务指令可以通过 `call.Context()` 读取本次渲染传入的根数据。推荐在控制器层构造一个规范的 `ViewContext`：

```go
type ViewContext struct {
    Site  SiteContext
    Route RouteContext
    Page  PageData
    Meta  PageMeta
}
```

推荐优先级：

```text
显式参数 > 页面配置 > 控制器写入 > 当前路径配置 > 全局系统配置
```

例如 `@sq:arc()` 未传 `cid` 时，可以从 `ViewContext.Page.CategoryID` 推导当前栏目；`@sq:arc(cid=99)` 则使用显式参数覆盖上下文。

## 9. 推荐下一步

建议继续看：

- [Support Matrix](support.md)
- [Business Tags](business-tags.md)
- [Examples](examples.md)
- [Integration Checklist](integration-checklist.md)
