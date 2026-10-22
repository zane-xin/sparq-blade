# Blade 示例操作文档

`examples/` 不只是零散 demo，而是一组可运行的操作文档。推荐先启动
`examples/docs`，它会提供一个 HTTP 文档站点，首页包含应用场景入口和导航，
各页面分别说明每种能力的实现方式。

## 启动文档站点

在 `sparq-blade` 目录下运行：

```bash
go run ./examples/docs
```

访问：

```text
http://127.0.0.1:8082/
http://127.0.0.1:8082/basic
http://127.0.0.1:8082/components
http://127.0.0.1:8082/data-tags
http://127.0.0.1:8082/request-context
http://127.0.0.1:8082/http
http://127.0.0.1:8082/healthz
```

## 首页内容

首页会说明这个库适合的应用场景：

- 后台管理系统页面渲染
- CMS 栏目页、文章列表页、详情页
- 插件系统中的模板渲染
- 需要业务标签 DSL 的页面
- 需要从当前访问路径推导数据的页面

首页还提供导航入口，点击后可以进入每个能力页面查看实现方式。

## 分页面说明

### 基础渲染

访问：

```text
http://127.0.0.1:8082/basic
```

说明：

- 如何创建 `Engine`
- 如何配置模板目录
- 如何调用 `RenderFile`
- 如何使用 `@extends`、`@section`、`@yield`
- 如何使用 `@include` 和 `@list`
- 如何在普通 `@list` 中使用 `loop.index`、`loop.first`、`loop.last`
- 如何使用 `@@sq:title` 输出字面量 `@sq:title`

对应示例目录：

```text
examples/basic
```

### 组件

访问：

```text
http://127.0.0.1:8082/components
```

说明：

- 如何使用 `@component`
- 如何使用 `@slot`
- 如何封装卡片、面板、列表行等复用 UI
- 如何注册自定义指令

对应示例目录：

```text
examples/components
```

### 数据标签

访问：

```text
http://127.0.0.1:8082/data-tags
```

说明：

- 如何注册 `blade.WithDataDirective`
- 如何在模板中使用 `@sq:arc(...) ... @end`
- 如何读取 `cid`、`limit` 等参数
- 如何返回 `blade.Items(...)`
- 如何在模板中使用 `field` / `field` 和 `loop` / `loop`

对应示例目录：

```text
examples/tags
examples/tags-best-practices
```

### 请求上下文

访问：

```text
http://127.0.0.1:8082/request-context
```

说明：

- 如何定义请求级 `ViewContext`
- 如何在 HTTP 请求中构造上下文
- 如何让 `@sq:title`、`@sq:seo` 从上下文读取页面元数据
- 如何让 `@sq:arc()` 从当前路径推导栏目
- 如何用模板参数 `cid=99` 覆盖上下文推导

对应示例目录：

```text
examples/context
```

### HTTP 集成

访问：

```text
http://127.0.0.1:8082/http
```

说明：

- 应用启动时复用一个 `Engine`
- 每次请求构造自己的 `ViewContext`
- handler 中调用 `RenderFile`
- 业务标签读取 `call.Context()`
- 避免使用包级全局变量保存请求数据

对应示例目录：

```text
examples/http
examples/docs
```

### CMS 组合场景

独立运行：

```bash
go run ./examples/cms
```

访问：

```text
http://127.0.0.1:8083/
http://127.0.0.1:8083/directives
http://127.0.0.1:8083/templates
http://127.0.0.1:8083/data
http://127.0.0.1:8083/articles
http://127.0.0.1:8083/navigation
http://127.0.0.1:8083/pagination
http://127.0.0.1:8083/custom
http://127.0.0.1:8083/healthz
```

说明：

- 使用 `LayeredFS` 实现主题覆盖、应用模板和模块默认模板
- 以中文文档站形式覆盖基础指令、模板组合、数据处理和自定义扩展
- 使用 `WithDataDirective("arc")` 注册数据型业务指令
- 使用 `WithTag("pager")` / `@sq:pager()` 输出分页结构
- 使用 `@children(..., wrap="ul")` 渲染直接子级菜单
- 使用 `WithDirective`、`WithBlockDirective` 演示自定义模板扩展
- 每个路由页面本身就是对应用法的 HTML 操作文档
- 模板里保持稳定路径，例如 `cms:partials.pager`

## 独立运行示例

如果只想运行某个示例，也可以直接执行：

```bash
go run ./examples/basic
go run ./examples/components
go run ./examples/tags
go run ./examples/tags-best-practices
go run ./examples/cms
go run ./examples/context
go run ./examples/http
```

`examples/context` 也是 HTTP 示例，访问地址：

```text
http://127.0.0.1:8081/news
http://127.0.0.1:8081/products
http://127.0.0.1:8081/healthz
```

## 模板指令覆盖

文档站点模板真实使用了这些指令：

- `@sq:title`
- `@sq:seo`
- `@extends`
- `@section`
- `@yield`
- `@include`
- `@component`
- `@slot`
- `@sq:arc(...) ... @end`
- `@list`
- `@if`

说明文字中如果需要展示字面量指令，可以使用 `@@` 转义。例如：

```blade
@@sq:title
@@sq:arc
```

渲染后会输出：

```text
@sq:title
@sq:arc
```

## 验证

运行：

```bash
go test ./...
go run ./examples/docs
```
