# Blade

使用 Go 实现的参考 Laravel Blade 的模板解析库。

当前版本已经具备一套可运行的最小能力集合，包含：

- Blade 基础语法解析与渲染
- 模板继承、组件、插槽、`@each`
- 自定义单指令与块级指令扩展
- 实验性业务指令 `@sq:arc(...) ... @end`、`@sq:pager(...)`
- 受控表达式函数、内建值转换和静态资源路径生成
- 模板文件解析、缓存与渲染测试

## Status

当前仓库处于“首个可用版本收尾”阶段：

- 已可用于字符串模板和文件模板渲染
- 已有基础文档、样例和测试
- API 正在收敛，暂未声明为稳定 `v1`
- 常用接入场景现在可只引入根包 `blade`

安装：

```bash
go get github.com/zane-xin/sparq-blade@latest
```

当前支持语法与已知限制见 [docs/support.md](docs/support.md)。
第一次接入建议先看 [QUICKSTART.md](QUICKSTART.md)。
CMS 项目接入前建议过一遍 [docs/integration-checklist.md](docs/integration-checklist.md)。
集中查阅文档可直接看 [docs/README.md](docs/README.md)。
公开 API 边界见 [docs/public-api.md](docs/public-api.md)。
对外变更记录见 [CHANGELOG.md](CHANGELOG.md)。

## Public API Boundary

稳定接入入口是根包：

```go
import blade "github.com/zane-xin/sparq-blade"
```

根包公开 `New`、`RenderString`、`Parse`、`Lex`、`WithFS`、`OSFS`、`LayeredFS`、`WithCache`、
`WithDirective`、`WithBlockDirective`、`WithTag`、`WithDataDirective` 以及业务标签辅助类型和函数。

底层解析器、渲染引擎、运行时和指令注册表实现位于 `internal/` 下，不作为稳定公开 API。

## Support Matrix

| 能力                       | 当前状态  | 说明                                                                                                           |
|--------------------------|-------|--------------------------------------------------------------------------------------------------------------|
| `{{ }}` / `{!! !!}` / 注释 | 已支持   | 普通输出自动按 HTML 上下文转义；原样输出仅用于应用生成或已可靠清洗的可信内容                                                                         |
| 条件指令                     | 已支持   | `@if`、`@elseif`、`@else`、`@unless`                                                                            |
| 循环指令                     | 已支持   | 流程指令直接使用 `@list`、`@children` 和统一的 `@end`；高阶场景支持 `@foreach`、范围式 `@for`                     |
| 模板继承                     | 已支持   | `@extends`、`@section`、`@yield`                                                                               |
| 局部模板引入                   | 已支持   | `@include('view')`、`@include('view', data)`                                                                 |
| 列表局部渲染                   | 已支持   | `@each`                                                                                                      |
| 组件与插槽                    | 已支持   | `@component('card')`、`@component('card', data)`、`@slot`                                                     |
| 模板路径解析                   | 已支持   | 支持 `default.view` 和 `cms:default.view`                                                                       |
| 编译型业务指令                 | 已支持   | Go 侧使用局部名注册，模板侧固定使用 `@sq:name`；支持单指令与块级指令                                                        |
| 业务标签                     | 实验性支持 | 数据型用 `WithDataDirective("arc")` + `@sq:arc(...) ... @end`；输出型用 `WithTag("pager")` + `@sq:pager(...)`                  |
| 表达式函数                    | 已支持   | 内建 `len`、`upper`、`formatDate`、`asset` 等函数，并支持 `WithFunction("cms.name", handler)` 注册命名空间函数                         |
| 表达式兼容性                   | 部分支持  | 当前是受控子集，不等价于完整 Blade/PHP 表达式；变量直接写 `field`，使用 `$field` 会返回语法错误                                                    |
| 组件参数协议                   | 部分支持  | 支持第二个参数合并到组件作用域，暂未实现完整属性/props 绑定模型                                                                          |
| 错误上下文                    | 已支持   | 编译/解析错误包含源码行、指针和文件模板名                                                                                        |
| 多阶段块指令                   | 未支持   | 暂不支持类似 `@if/@elseif/@else` 的自定义块协议                                                                           |

## Examples

仓库内提供了可直接运行的样例，见 [examples/README.md](examples/README.md)。

常用命令：

```bash
go run ./examples/basic
go run ./examples/components
go run ./examples/tags
go run ./examples/tags-best-practices
go run ./examples/http
```

建议起步顺序：

1. 先跑 `./examples/basic` 熟悉继承、区块和 include
2. 再跑 `./examples/components` 看组件、插槽和自定义指令
3. 再跑 `./examples/tags` 看 `@sq`、`as=article`、`@children` 和 `out`
4. 再跑 `./examples/tags-best-practices` 看 CMS 业务标签约定、分页和空结果处理
5. 如果是 Web 项目，再跑 `./examples/http`

## Business Tag Rule

业务指令统一使用 `sq:` 品牌入口，流程指令不使用品牌前缀：

- 取数据并循环：`blade.WithDataDirective("arc", handler)`，模板写 `@sq:arc(..., as=article) ... @end`
- 直接输出 HTML：`blade.WithTag("pager", handler)`，模板写 `@sq:pager(...)`
- 编译期单指令：`blade.WithDirective("badge", handler)`，模板写 `@sq:badge(...)`
- 编译期块指令：`blade.WithBlockDirective("panel", "", handler)`，模板写 `@sq:panel(...) ... @end`
- 列表和子级结构：优先使用 `@list(...)`、`@children(...)`

数据业务块会直接遍历 handler 返回的 `Items`，块内只写单项模板，不需要再嵌套 `@list`。只有使用 `out=articles` 导出数据后，才通过 `@list(articles, as=article)` 继续处理。

所有外部注册指令都按业务局部名接收，调用方不能传入 `sq:`、`pt:` 或其他命名空间。局部名也不能与 `list`、`if`、`children` 等内建流程指令重名。

## Verification

当前仓库最近一次本机验证命令：

```bash
GOOS=darwin GOARCH=amd64 GOCACHE=/tmp/blade-gocache go test ./...
```

## 单仓构建与发布

当前目录可独立建立 Git 仓库，Go module 路径保持不变。`make test` 和 `make build` 可用于本地开发；`make check-release` 始终使用 `GOWORK=off` 验证正式依赖、测试、静态检查和构建。

```sh
make check-release
make release VERSION=0.1.0
```

依赖中的 `v0.0.0` 是尚未发布的占位版本，需要在依赖仓库发布真实 tag 后执行 `GOWORK=off go get <module>@<version>` 更新 go.mod/go.sum。本地 replace 不允许进入正式发布。源码库的 release 仅执行发布门禁，之后由维护者提交并创建版本 tag；sparq-tool 额外生成包含 CLI 与 Updater 的平台归档及 SHA256。
