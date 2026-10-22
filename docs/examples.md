# 示例操作文档

推荐先启动示例文档站点：

```bash
go run ./examples/docs
```

访问首页：

```text
http://127.0.0.1:8082/
```

首页会展示项目应用场景入口和导航。每个导航页面都是一个独立操作文档，
对应说明一种模板能力。

## 页面导航

| 页面 | 地址 | 说明 |
| --- | --- | --- |
| 首页 | `http://127.0.0.1:8082/` | 应用场景入口和推荐顺序 |
| 基础渲染 | `http://127.0.0.1:8082/basic` | `Engine`、`RenderFile`、继承、include、`@list` |
| 组件 | `http://127.0.0.1:8082/components` | `@component`、`@slot`、自定义指令 |
| 数据标签 | `http://127.0.0.1:8082/data-tags` | `WithDataDirective`、`@sq:arc`、`field`、`loop` |
| 请求上下文 | `http://127.0.0.1:8082/request-context` | `call.Context()`、`@sq:title`、`@sq:seo`、自动推导栏目 |
| 语法定制 | `http://127.0.0.1:8082/syntax` | 使用 `blade.WithSyntax` 定制输出、原样输出和注释分隔符 |
| HTTP 集成 | `http://127.0.0.1:8082/http` | Web 项目中复用 `Engine` |
| CMS 中文文档站 | `http://127.0.0.1:8083/` | 基础指令、模板组合、数据处理、自定义扩展、`LayeredFS`、业务标签 |
| 健康检查 | `http://127.0.0.1:8082/healthz` | 返回 `ok` |

## 模板指令应用

文档站点模板会真实使用这些指令：

```blade
<title>@sq:title</title>
@sq:seo

@extends('layout')
@section('content')
  @include('partials.doc-note', Page.IncludeData)

  @component('components.doc-card', Page.CardData)
    组件内容
  @endcomponent

  @sq:arc(limit=2, as=article)
    <li>{{ article.title }}</li>
  @end
@endsection
```

如果页面正文需要展示字面量指令，可以使用 `@@` 转义：

```blade
@@sq:title
@@sq:arc
```

普通 `@list` 内可以读取 `loop`：

```blade
@list(items, as=item)
  {{ loop.index }} - {{ item.title }}
@empty
  暂无内容
@endlist
```

## 模板组合传参

`@include` 和 `@component` 都可以接收第二参数：

```blade
@include('partials.note', note)

@component('components.card', card)
  默认 slot 内容
@endcomponent
```

第二参数会展开为局部变量，同时保留根 `ViewContext`。因此局部模板既能读取 `title` / `title`、`body` / `body` 这类传入数据，也能继续读取 `Page.Title` / `Page.Title`。

`@extends` 不单独传参。布局页、页面页和局部模板共享同一个根 `ViewContext`，布局需要的站点、SEO、导航等数据应在控制器层写入根上下文。

## 请求上下文场景

访问：

```text
http://127.0.0.1:8082/request-context
```

这个页面演示：

- 路由层解析当前访问路径
- 控制器构造 `ViewContext`
- `@sq:title` 和 `@sq:seo` 读取 `ViewContext.Meta`
- `@sq:arc()` 未传 `cid` 时读取 `ViewContext.Page.CategoryID`
- `@sq:arc(cid=99)` 使用显式参数覆盖上下文

## 独立示例

除文档站点外，也可以直接运行独立示例：

```bash
go run ./examples/basic
go run ./examples/components
go run ./examples/tags
go run ./examples/tags-best-practices
go run ./examples/context
go run ./examples/http
```

`examples/context` 也是 HTTP 示例：

```text
http://127.0.0.1:8081/news
http://127.0.0.1:8081/products
http://127.0.0.1:8081/healthz
```
