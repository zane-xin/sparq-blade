# 当前支持语法与限制

本文档用于描述当前 `blade` 库已经支持的语法能力、扩展方式与已知限制。

## 一、当前支持的核心语法

### 1. 输出与注释

- `{{ var }}`
  - 自动 HTML 转义输出
- `{!! var !!}`
  - 原样输出，不做 HTML 转义
- `{{-- comment --}}`
  - 模板注释，不进入最终输出

### 2. 条件指令

- `@if(...)`
- `@elseif(...)`
- `@else`
- `@endif`
- `@unless(...)`
- `@endunless`

当前表达式已支持：

- 变量路径，例如 `user.name`
- 字符串、数字、布尔值、`nil` / `null`
- `==`、`!=`
- `>`、`>=`、`<`、`<=`
- `&&`、`||`
- `!`
- 括号分组

### 3. 循环指令

- `@foreach(items as item)`
- `@foreach(items as key => item)`
- `@endforeach`
- `@for(init; condition; step)`
- `@endfor`
- `@while(condition)`
- `@endwhile`

说明：

- `@foreach` 可用
- `@for` 可用，已支持简单赋值、自增、自减、`+=`、`-=`
- `@while` 具备安全上限，避免无限循环

### 4. 模板组合

- `@include('view')`
- `@include('view', data)`
- `@extends('layout')`
- `@section('content')`
- `@endsection`
- `@yield('content')`
- `@each('row', items, 'item', 'empty')`

说明：

- 模板文件名支持不带后缀调用
- 模板名支持点号路径，例如 `default.view` -> `default/view`
- 模板名支持可选命名空间前缀，例如 `cms:default.view`
- `@include` 的第二个参数会合并到被引入模板的作用域中
- 默认可解析后缀：
  - `.blade.php`
  - `.blade`
  - `.tmpl`
  - `.html`

### 5. 组件与插槽

- `@component('card') ... @endcomponent`
- `@component('card', data) ... @endcomponent`
- `@slot('title') ... @endslot`

说明：

- 默认插槽通过组件模板中的 `slot` 读取
- 命名插槽通过组件模板中的 `title`、`footer` 这类变量读取
- `@component` 的第二个参数会合并到组件模板的作用域中
- 插槽内容仍按调用处作用域渲染

### 6. `@go`

支持两种形式：

- 单指令形式：
  - `@go(printf "%s-%d" .name .count)`
- 块形式：
  - `@go ... @endgo`

说明：

- `@go(...)` 会编译成原生 Go 模板 action
- `@go ... @endgo` 内部内容按原生 Go 模板源码原样注入，不再继续按 Blade 语法解析
- `@go` 内部当前 `.` 可直接访问根数据，例如 `.name`
- 同时保留 `Root` 字段，因此 `.Root.name` 也仍然可用

### 7. 业务标签

支持实验性业务标签：

- `@sq('cms:arc', limit=1)`
- `@sq('cms:list', limit=10) ... @endsq`

说明：

- Go 侧通过 `blade.WithDataTag("cms:arc", handler)` 这类方式按目标注册
- 模板侧入口固定为 `@sq`
- 第一参数必须是字符串字面量目标，例如 `'cms:arc'`
- 后续参数当前支持 `key=value` 命名参数，`out` 额外支持裸参数简写
- 参数值当前支持字符串、数字、布尔、`nil` / `null`、变量表达式
- `id=field` 控制块模式下当前项变量名，默认是 `field`
- `out` 默认导出到 `field`
- `out='articleList'` 可显式指定导出变量名
- 单指令 `out` 可导出列表结果或单对象结果
- 块标签当前只支持 `Items` 列表结果，不支持单对象 `Value`
- 块标签会自动遍历结果集，并注入 `field` 与 `loop.index`、`loop.iteration`、`loop.count`、`loop.remaining`、`loop.first`、`loop.last`、`loop.odd`、`loop.even`
- 单对象标签返回零值 `DataTagResult{}` 时，模板侧 `out` 变量按空值处理，可直接用 `@if(field)` 判空
- 列表标签块模式返回空 `Items` 时，不输出任何内容，也不会报错

## 二、自定义指令扩展

### 1. 单指令扩展

可通过公开 API 注册：

```go
engine := blade.New(
    blade.WithDirective("badge", func(expression string) (string, error) {
        return fmt.Sprintf(`<strong>{{ bladeEcho . %q }}</strong>`, expression), nil
    }),
)
```

模板中可使用：

```blade
@badge(name)
```

### 2. 块级指令扩展

可通过公开 API 注册：

```go
engine := blade.New(
    blade.WithBlockDirective("panel", "", func(expression string, body string) (string, error) {
        return `<section class="panel">` + body + `</section>`, nil
    }),
)
```

模板中可使用：

```blade
@panel()
  Hello {{ name }}
@endpanel
```

说明：

- 当 `endName` 为空时，默认结束指令为 `end` + 指令名
- 目前自定义块指令只支持“单层包裹式”编译，不支持更复杂的多分支块协议

## 三、当前公开 API

当前可直接使用的入口包括：

- `blade.Lex`
- `blade.Parse`
- `blade.New`
- `blade.RenderString`
- `blade.WithDirective`
- `blade.WithBlockDirective`
- `blade.WithDataTag`
- `blade.WithTag`

文件系统与缓存配置当前仍从 `engine` 包注入，例如：

- `blade.WithFS`
- `blade.WithCache`
- `blade.OSFS`

如需命名空间模板目录，可使用：

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

## 四、错误提示

编译和解析阶段的结构化错误会包含：

- 错误码
- 行列号
- 错误消息
- 源码行
- 指针位置
- 文件模板场景下的模板名

## 五、已知限制

以下内容当前尚未覆盖或只做了最小实现：

- `@while` 仍不适合写依赖模板体内状态变化的复杂循环逻辑
- 自定义指令目前不支持“类似 `@if/@elseif/@else` 的多阶段块协议”
- 业务标签当前仍是实验性 API，语法和处理协议后续仍可能微调
- 业务标签当前支持列表结果和单对象结果，但仍不支持对象参数、数组参数、查询串参数
- 组件当前只支持视图组件，不支持更完整的参数绑定协议
- `@go ... @endgo` 内部不会继续解析 Blade 语法，因此不应混写 Blade 指令
- 表达式解析目前仍是受控子集，不支持完整 Go / PHP / Blade 表达式语言
- 文件缓存当前基于模板内容版本，不包含更高级的依赖图级失效策略

## 六、建议使用方式

对于当前版本，推荐这样使用：

1. 页面主模板优先使用 Blade 语法组织结构
2. 只在确实需要时局部使用 `@go`
3. 组件和 `@each` 适合承载稳定的复用片段
4. 自定义指令优先实现简单包装型语义，不要一开始引入复杂 DSL
