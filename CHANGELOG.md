# Changelog

本文件记录 `blade` 的对外可见变更。

## Unreleased

当前阶段为首个可用版本收尾，尚未打正式版本 tag。

### Added

- 提供根包公开入口：`blade.New`、`blade.RenderString`、`blade.Lex`、`blade.Parse`
- 提供根包常用配置别名与选项：`blade.WithFS`、`blade.WithCache`、`blade.WithDirectives`、`blade.OSFS`
- 实现编译链路：Lexer、Parser、AST、Generator
- 实现字符串模板与文件模板渲染
- 支持模板继承：`@extends`、`@section`、`@yield`
- 支持模板引入：`@include`
- 支持 `@include('view', data)` 显式数据传入
- 支持条件指令：`@if`、`@elseif`、`@else`、`@unless`
- 支持循环指令：`@list`、`@children`、`@foreach`、范围式 `@for(i in 1..5)`
- 支持输出与注释：`{{ }}`、`{!! !!}`、`{{-- --}}`
- 支持复用能力：`@each`、`@component`、`@slot`
- 支持 `@component('view', data)` 显式数据传入
- 支持通过固定 `sq:` 品牌入口访问编译型单指令与块级业务指令
- 增加输出型业务标签 API：`blade.WithTag("pager")`，模板侧使用 `@sq:pager(...)`
- 增加实验性数据标签 API：`blade.WithDataDirective("arc")`、`as=field`、`out`、`out='arc'`
- 增加数据标签返回辅助：`blade.Rows`、`blade.Object`，支持 map 与 struct
- 裸 `out` 默认导出到 `field`
- 数据标签单指令 `out` 支持导出单对象结果
- 数据标签块模式明确仅支持 `Items` 列表结果
- 数据标签块直接遍历 `Items`，并支持通过 `@empty` 处理空结果
- 为数据标签块增加 `loop.remainini`、`loop.odd`、`loop.even`
- 模板变量路径使用 `field` 或 `article.title` 形式
- 支持受控表达式函数、嵌套调用和插件命名空间函数
- 增加 `len`、`upper`、`lower`、`contains`、`default`、`coalesce`、`formatDate`、`asset` 内建函数
- 增加 `WithFunction`、`ValidateFunction` 和 `WithAssetPrefix` 根包 API
- 提供 `blade.OSFS` 以及有界 AST、编译文本和文件模板缓存
- 限定默认文件模板后缀为 `.tpl`、`.html`
- 增强本地文件加载安全，拒绝 path traversal 和 symlink escape
- 将底层 compiler、engine、runtime、directives 实现迁入 `internal`
- 增加真实页面级组合测试，覆盖命名空间、点号路径、include/component 数据传入、each、slot 和自定义块指令
- 增加 `basic`、`components`、`context`、`docs`、`http`、`tags` 运行样例
- 增加 Quick Start、支持说明、公开 API 边界和接入检查文档
- 增加公开 API 与包级 Go doc
- 增强编译/解析错误信息，包含源码行、指针位置和文件模板名
- 增加 fuzz tests 与渲染 golden tests

### Changed

- 业务指令前缀从 `@pt:` 调整为 `@sq:`，旧前缀不再注册兼容别名
- `sq:` 仅用于业务指令；`@if`、`@list`、`@children`、`@empty`、`@end` 等流程指令不提供品牌别名
- `WithDirective`、`WithBlockDirective`、`WithTag`、`WithDataDirective` 统一只接收局部名，并映射到内部固定的 `sq:` 入口
- 模板表达式不再接受 `$` 变量前缀；输出、条件、循环和指令参数统一直接写变量名

### Fixed

- 修复扩展处理函数为空时延迟到渲染阶段 panic 的问题
- 隔离注入指令注册表和公开 AST 的可变状态，避免并发竞态与缓存污染
- 修复输出标签 `as` 参数误判、业务块内 section 替换、逻辑运算短路和大整数比较
- OSFS 使用内容摘要生成版本，不再保留进程级无界源码缓存

### Known Limits

- 表达式语言当前是受控子集，不是完整 Blade / PHP 表达式兼容层
- 组件当前偏视图复用，不包含完整属性绑定协议
- 自定义块指令当前只支持单层包裹式协议

### Verification

- `GOOS=darwin GOARCH=amd64 GOCACHE=/tmp/blade-gocache go test ./...`
