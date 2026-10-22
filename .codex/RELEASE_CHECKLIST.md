# 发布检查清单

本文档记录 `blade` 当前版本发布前需要确认的事项。

## 一、当前 API 边界

推荐对外入口：

- `blade.RenderString`
- `blade.New`
- `blade.WithFS`
- `blade.WithCache`
- `blade.WithDirective`
- `blade.WithBlockDirective`
- `blade.OSFS`

当前已纳入推荐能力：

- 点号路径模板名，例如 `pages.home`
- 命名空间模板名，例如 `cms:layouts.default`
- `@include('view', data)`
- `@component('view', data)`
- 编译/解析错误源码上下文

高级或调试入口：

- `blade.Lex`
- `blade.Parse`
- `compiler` 包
- `engine` 包

内部运行时入口：

- `runtime` 包主要服务于编译后的模板执行，不建议业务代码直接依赖

## 二、发布前验证命令

当前本机 shell 存在交叉编译环境变量，验证时建议显式指定本机目标：

```bash
GOOS=darwin GOARCH=amd64 GOCACHE=/tmp/blade-gocache go test ./...
GOOS=darwin GOARCH=amd64 GOCACHE=/tmp/blade-gocache go run ./examples/basic
GOOS=darwin GOARCH=amd64 GOCACHE=/tmp/blade-gocache go run ./examples/components
GOOS=darwin GOARCH=amd64 GOCACHE=/tmp/blade-gocache go run ./examples/tags
GOOS=darwin GOARCH=amd64 GOCACHE=/tmp/blade-gocache go run ./examples/tags-best-practices
GOOS=darwin GOARCH=amd64 GOCACHE=/tmp/blade-gocache go run ./examples/http
```

HTTP 示例启动后验证：

```bash
curl -sS http://127.0.0.1:8080/healthz
curl -sS http://127.0.0.1:8080/
curl -sS http://127.0.0.1:8080/users
```

## 三、已通过验收

- 完整测试通过
- `examples/basic` 运行通过
- `examples/components` 运行通过
- `examples/tags` 运行通过
- `examples/tags-best-practices` 运行通过
- `examples/http` 启动通过
- `/healthz` 返回 `ok`
- `/` 渲染首页正常
- `/users` 渲染用户列表正常
- 真实页面级组合测试通过，覆盖命名空间、点号路径、include/component 数据传入、each、slot、`@go` 和自定义块指令

## 四、v0.2.0 已完成能力

- 模板命名空间读取：`cms:default.view`
- 点号路径读取：`default.view`
- `@include('view', data)` 数据传入
- `@component('view', data)` 数据传入
- 编译/解析错误上下文
- 真实页面级组合测试

## 五、tag 前人工确认

- 确认 `README.md`、`QUICKSTART.md`、`CHANGELOG.md` 描述一致
- 确认 `.codex/INTEGRATION_CHECKLIST.md` 与当前 `@sq` 业务标签协议一致
- 确认是否要清理外层工作区中与本包无关的改动
- 确认版本号策略，例如 `v0.2.0`
- 确认是否需要在发布前固定 Go 版本或 CI 环境
- 确认是否接受 `compiler`、`engine`、`directives` 子包作为高级 API 暴露
