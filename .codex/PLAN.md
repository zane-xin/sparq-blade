# Blade 当前计划

本文档记录 `blade` 当前实现状态、已完成能力和下一阶段任务。具体语法支持范围见 [SUPPORT.md](SUPPORT.md)。

## 一、当前状态

项目已经从初始化阶段进入可用版本收尾阶段，核心链路已经打通：

- 已建立 `compiler/`、`engine/`、`runtime/`、`directives/` 等核心包
- 已提供根包公开入口 `blade.go` 与包文档 `doc.go`
- 已实现从 Blade 模板源码到 Go template 的编译链路
- 已实现字符串模板与文件模板渲染
- 已有覆盖编译器、引擎、公开 API、示例和基准测试的测试集
- 已提供 `basic`、`components`、`http` 三套可运行样例
- 已补充 `README.md`、`QUICKSTART.md`、`examples/README.md` 等接入文档

最近一次验证命令：

```bash
GOOS=darwin GOARCH=amd64 GOCACHE=/tmp/blade-gocache go test ./...
```

说明：当前本机 shell 里存在 `GOOS=linux GOARCH=amd64` 环境，会导致 macOS 下测试二进制 `exec format error`，本地验证时需显式指定本机目标或清理相关环境变量。

## 二、已完成能力

### 1. 编译器

- Lexer 支持文本、输出、原样输出、注释、普通指令、`@go` 块
- Parser 支持嵌套结构、闭合校验和结构化错误
- AST 已覆盖当前核心语法节点
- Generator 可生成可执行的 Go template 源码

### 2. 模板语法

- 输出：`{{ }}`、`{!! !!}`、`{{-- --}}`
- 条件：`@if`、`@elseif`、`@else`、`@unless`
- 循环：`@foreach`、`@for`、`@while`
- 模板组合：`@include`、`@extends`、`@section`、`@yield`
- 复用：`@component`、`@slot`、`@each`
- Go template 逃生口：`@go(...)`、`@go ... @endgo`
- 自定义指令：单指令与块级指令

### 3. 引擎层

- `blade.RenderString`
- `blade.New`
- `RenderString`
- `RenderFile`
- 文件系统抽象 `engine.OSFS`
- 编译缓存与文件读取缓存
- 模板后缀解析：`.blade.php`、`.blade`、`.tmpl`、`.html`

### 4. 示例与文档

- `examples/basic` 展示继承、区块、include、foreach 和输出
- `examples/components` 展示组件、插槽、`@each`、`@go` 和自定义指令
- `examples/http` 展示 `net/http` 集成
- `QUICKSTART.md` 已覆盖安装、字符串渲染、文件渲染、组件、`@go`、自定义指令和 HTTP 接入

## 三、当前限制

这些限制不阻塞第一版可用，但需要在发布说明中明确：

- 表达式语言是受控子集，不等价于完整 PHP Blade 或 Go 表达式
- `@while` 带安全上限，不适合依赖模板体内状态变化的复杂循环
- 组件当前偏视图复用，不支持完整属性绑定协议
- 自定义块指令只支持单层包裹式转换，不支持多阶段协议
- `@go ... @endgo` 内部不会再解析 Blade 指令
- 文件缓存当前基于源内容和解析结果，不包含完整依赖图失效策略

## 四、后续版本规划

### v0.1.x：首个可用版本收尾

目标：固定当前边界，准备第一个可用 tag。

当前已完成：

- 根包 API 收敛：`blade.WithFS`、`blade.WithCache`、`blade.OSFS`
- 包级 Go doc 与公开 API 注释
- `README.md`、`QUICKSTART.md`、`CHANGELOG.md`
- 三套 example 运行验证
- 发布检查清单 `.codex/RELEASE_CHECKLIST.md`

剩余人工事项：

- 确认版本号策略，例如 `v0.1.0`
- 确认是否接受 `compiler`、`engine`、`directives` 子包作为高级 API 暴露
- 确认是否需要 CI / 固定 Go 版本后再打 tag

### v0.2.0：模板组织与数据传入增强

目标：优先提升真实项目可接入性，不扩大自定义指令语法复杂度。

任务：

- 支持模板命名空间读取，例如 `@extends('cms:default.view')`
- 为 `@extends`、`@include`、`@component`、`@each` 统一补充命名空间模板查找测试
- 明确并测试点号路径读取，例如 `@extends('default.view')`
- 支持 `@include('view', data)` 或等价数据传入方式
- 支持 `@component('card', data)` 或等价参数传入方式
- 增强错误信息，至少包含源码片段、错误行和定位指针
- 增加更接近真实页面的组合测试
- 同步更新 `README.md`、`QUICKSTART.md`、`.codex/SUPPORT.md`、`CHANGELOG.md`

当前已完成：

- 模板命名空间读取
- 点号路径读取测试
- `@include('view', data)`
- `@component('card', data)`
- 编译/解析错误上下文
- 真实页面级组合测试

当前取舍：

- 模板命名空间优先做在文件系统/模板解析层，不直接耦合到指令语法
- 点号路径 `default.view` 当前已有基础支持，下个版本重点是补测试和文档，不是重做语法

### v0.3.0：业务标签与扩展机制评估

目标：在不破坏当前解析稳定性的前提下，评估更复杂的业务标签场景。

设计文档：

- [BUSINESS_DIRECTIVES.md](BUSINESS_DIRECTIVES.md)

任务：

- 实现“保守版标签指令”语法，例如 `@sq('cms:arc', limit=1, pid=1)`
- 明确单指令与块指令的统一协议，避免引入 `@sq:cms:arc ... @end:cms::arc` 这类高复杂度闭合语法
- 实现第一版参数解析策略，支持字符串、数字、布尔、nil、变量表达式
- 增加 `TagCall`、`TagFunc`、`WithTag` 这类业务标签 API
- 明确自定义指令生成 Go template 片段时的安全边界
- 增加单指令、块指令、错误路径测试

当前结论：

- 暂不优先支持带冒号的原生命名空间指令名
- 优先保留普通指令名，把业务目标写入参数，例如 `@sq('cms:arc', limit=1, pid=1)`
- 如果后续需要块形式，优先考虑 `@sq(...) ... @endsq` 这类统一闭合风格
- 暂不支持对象参数、数组参数、查询串参数、多阶段标签块

当前已完成：

- `TagCall`、`TagFunc`、`WithTag`
- `DataTagCall`、`DataTagFunc`、`WithDataTag`
- `@sq('cms:arc', limit=1)` 单指令
- `@sq(...) ... @endsq` 块指令
- 第一版命名参数解析
- `id=field` / `out` / `out='arc'`
- 裸 `out` 默认导出到 `field`
- 单指令 `out` 支持导出单对象结果
- 单指令、块指令、错误路径测试

### v0.4.0：性能与缓存

目标：在功能边界稳定后再系统优化性能。

任务：

- 重新建立基准测试基线
- 评估 include、extends、component 的依赖图缓存
- 评估文件变更检测与缓存失效策略
- 减少渲染路径中的临时 map 合并和重复解析
- 根据真实项目模板再决定是否优化 lexer/parser 分配

## 五、建议执行顺序

当前建议按版本顺序推进：

1. 完成 `v0.1.x` 的人工发布确认
2. 进入 `v0.2.0`，先做模板命名空间
3. 再做 `@include` / `@component` 数据传入
4. 然后补错误上下文和真实组合测试
5. `v0.3.0` 再讨论保守版业务标签指令的最终语法
