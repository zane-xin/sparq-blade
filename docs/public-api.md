# Public API

`sparq-blade` 的稳定接入入口是根包：

```go
import blade "github.com/zane-xin/sparq-blade"
```

## Stable Surface

当前对外建议依赖这些根包 API：

- 渲染入口：`New`、`RenderString`
- 解析调试入口：`Lex`、`Parse`
- 引擎配置：`WithFS`、`WithCache`、`WithSyntax`、`WithAssetPrefix`
- 表达式函数：`Function`、`WithFunction`、`ValidateFunction`
- 文件系统：`OSFS`、`LayeredFS`、`TemplateFS`
- 缓存协议：`Cache`
- 编译型业务指令：`WithDirective`、`WithBlockDirective`
- 自定义输出型业务标签：`WithTag`
- 输出型业务标签调用：`TagCall`、`TagFunc`

## Experimental Surface

以下 API 当前可用，但仍按实验性能力处理，v1 前可能调整：

- `WithDataDirective`
- `DataTagCall`
- `DataTagResult`
- `DataFields`
- `Fields`
- `Item`
- `Items`
- `Rows`
- `ValueOf`
- `Object`
- `EmptyResult`

## Expression Functions

`WithFunction(name, handler)` 注册受控表达式函数，handler 固定接收 `args ...any` 并返回 `(any, error)`。函数可以用于输出、条件、循环集合和业务标签参数。

`asset`、`coalesce`、`contains`、`default`、`formatDate`、`len`、`lower`、`upper` 是内建函数，不能覆盖。插件函数使用带点号的命名空间，例如 `cms.formatDate`。

`WithAssetPrefix` 只配置 `asset(path)` 生成的 URL 路径前缀，不挂载静态目录，也不读取文件。静态目录仍由 HTTP Router 配置。

函数只承担纯值转换。需要读取数据库、网络、文件或输出可信 HTML 时，应使用 ViewContext、服务、`WithTag` 或 `WithDataDirective`。

## Business Directive Names

`WithDirective`、`WithBlockDirective`、`WithTag`、`WithDataDirective` 都只接受业务局部名。模板侧统一使用内部固定的 `sq:` 品牌入口；`@if`、`@list`、`@empty`、`@end` 等内建流程指令不使用品牌前缀。

显式传入 `sq:name`、`pt:name` 或其他命名空间会在 Engine 初始化阶段失败。

## Business Tag Helpers

`WithTag` 适合注册直接输出 HTML 的业务标签，例如 `@sq:pager()`、`@sq:crumb()`。

业务标签用业务名注册，例如 `WithTag("pager", handler)`；模板侧固定使用品牌入口 `@sq:pager()`，不会同时注册裸 `@pager()`。

`TagCall` 提供：

- `Context()`：读取当前渲染根上下文，用于默认推导分页、面包屑等业务数据
- `Has`、`Get`：读取原始参数
- `GetStringDefault`、`GetIntDefault`、`GetBoolDefault`：读取带默认值的类型化参数
- `Render(view, data)`：用当前引擎渲染一个模板文件，并继承当前渲染作用域

## Internal Packages

底层实现位于 `internal/` 下：

- `internal/compiler`
- `internal/engine`
- `internal/runtime`
- `internal/directives`

这些包不作为稳定公开 API，外部项目不应直接依赖。

## Compatibility

在 `v1` 发布前，根包 API 仍可能根据真实接入反馈做收敛调整。`v1` 后按语义化版本维护兼容性。
