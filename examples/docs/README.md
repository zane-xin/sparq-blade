# 示例操作文档

这个示例会启动一个 HTTP 服务，用当前 blade 库渲染一组中文操作文档页面。

```bash
go run ./examples/docs
```

访问地址：

```text
http://127.0.0.1:8082/
```

页面入口：

| 页面 | 地址 | 说明 |
| --- | --- | --- |
| 首页 | `http://127.0.0.1:8082/` | 项目应用场景入口、导航和流程总览 |
| 基础渲染 | `http://127.0.0.1:8082/basic` | `Engine`、`RenderFile`、layout、include、list |
| 组件与插槽 | `http://127.0.0.1:8082/components` | `@component`、`@slot`、可复用模板片段 |
| 数据标签 | `http://127.0.0.1:8082/data-tags` | `WithDataDirective`、`@sq:arc`、`field`、`loop` |
| 请求上下文 | `http://127.0.0.1:8082/request-context` | `call.Context()`、`@sq:title`、`@sq:seo`、路径推导数据 |
| 语法定制 | `http://127.0.0.1:8082/syntax` | 使用 `blade.WithSyntax` 定制输出、原样输出和注释分隔符 |
| HTTP 服务 | `http://127.0.0.1:8082/http` | 在 `net/http` 中复用 `Engine` |
| 健康检查 | `http://127.0.0.1:8082/healthz` | 返回 `ok` |

示例模板会实际使用 `@sq:title`、`@sq:seo`、`@include`、`@component`、`@slot`、`@sq`、`@list`、`@if` 等指令。

`/basic` 页面演示 `@include('partials.doc-note', Page.IncludeData)`，局部模板会读取传入数据并继续访问根上下文。

`/components` 页面演示 `@component('components.doc-card', Page.CardData)`，组件模板会读取传入数据，并渲染默认 slot 内容。

语法定制示例：

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

对应模板：

```blade
[[ title ]]
[[! html !]]
[[-- 这段不会输出 --]]
```
