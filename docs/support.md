# Support Matrix

本文档描述当前 `blade` 已支持的语法能力、扩展方式和已知限制。

## 1. 当前已支持

### 输出与注释

- `{{ var }}`
- `{!! var !!}`
- `{{-- comment --}}`

`{{ }}` 会根据 HTML 上下文自动转义输出。`{!! !!}` 会绕过转义，只能输出由应用生成或经过可靠 HTML 清洗的可信内容；不得直接输出请求参数、数据库富文本或其他外部输入。

### 条件指令

- `@if`
- `@elseif`
- `@else`
- `@unless`

表达式当前支持：

- 变量路径直接写 `user.name`；`$user.name` 会返回语法错误
- 字符串、数字、布尔、`nil` / `null`
- `==`、`!=`
- `>`、`>=`、`<`、`<=`
- `&&`、`||`
- `!`
- 括号分组

函数调用属于同一套受控表达式，可以用于转义输出、原样输出、条件、循环集合、指令参数和业务标签参数，也支持嵌套：

```blade
{{ upper(user.name) }}
{{ formatDate(article.createdAt, "2006-01-02") }}

@if(contains(user.roles, "admin"))
  管理员
@endif
```

当前内建函数：

- `len(value)`：返回字符串的 Unicode 字符数，或数组、切片、Map 的元素数。
- `upper(value)`、`lower(value)`：转换字符串大小写。
- `contains(value, target)`：检查字符串是否包含子串、数组或切片是否包含值、Map 是否包含键。
- `default(value, fallback)`：value 为模板假值时返回 fallback。
- `coalesce(values...)`：返回第一个非 nil 值。
- `formatDate(value, layout)`：使用 Go layout 格式化 `time.Time` 或 `*time.Time`，保留值自身时区；nil 返回空字符串。
- `asset(path)`：在站内相对路径前补充静态资源前缀，默认前缀为 `/assets`。

应用可以注册纯值转换函数：

```go
engine := blade.New(
    blade.WithFunction("cms.label", func(args ...any) (any, error) {
        if len(args) != 1 {
            return nil, fmt.Errorf("cms.label 需要 1 个参数")
        }
        return "[" + fmt.Sprint(args[0]) + "]", nil
    }),
    blade.WithAssetPrefix("/static"),
)
```

内建函数不能覆盖，重复函数名会在引擎创建时失败。函数不用于数据库、网络、文件访问、业务查询或 HTML 渲染；这些行为应放在 ViewContext、服务、`WithTag` 或 `WithDataDirective` 中。函数返回 Go 模板的可信 HTML、URL、CSS 或 JavaScript 类型会被拒绝。

### 循环指令

- `@list(items, as=item)`
- `@children(as=child)`
- `@children(as=child, wrap="ul", class="submenu")`
- `@empty`
- `@end`
- `@foreach(items as item)`
- `@for(i in 1..5)`

说明：

- 普通列表模板推荐使用 `@list`，并用 `@empty` 描述空列表内容
- 当前项的直接子级推荐使用 `@children`，默认读取 `children` 字段，不自动递归
- `@children` 的 `wrap` 只接受简单标签名，有子级时才输出包裹标签；`class` 可为该包裹标签追加 class
- `@list`、`@children` 和数据业务块推荐使用统一结束标签 `@end`
- 条件、传统循环和模板组合仍支持对应的专用结束标签，例如 `@endif`、`@endforeach`、`@endsection`
- 流程指令不提供 `sq:` 品牌别名；所有外部注册指令都通过固定的 `sq:` 业务入口访问
- `@foreach` 保留为高阶遍历能力，参数直接写变量名
- `@for` 仅支持整数范围，例如 `@for(i in 1..5)` 和 `@for(i in 5..1)`
- `@for`、`@foreach`、`@list`、`@children`、`@each` 和数据标签块单次最多迭代 10,000 项
- `@while` 不支持

### 模板组合

- `@include('view')`
- `@include('view', data)`
- `@extends`
- `@section`
- `@yield`
- `@each`

模板继承和运行时组合链最多允许 64 层；出现继承循环、include/component/each 自引用或间接循环时，渲染会返回包含调用链的错误。

### 组件与插槽

- `@component('card')`
- `@component('card', data)`
- `@slot('title')`

### 自定义业务指令

- `blade.WithDirective`
- `blade.WithBlockDirective`
- 模板入口固定为 `@sq:name`，块指令推荐使用内建 `@end` 结束

### 业务标签

- `blade.WithDataDirective`
- `blade.WithTag`
- `@sq:arc(...)`
- `@end`
- `@sq:pager(...)`
- `as=field`
- `out`
- `out=name`
- `@children`
- `Rows` / `Object`
- `Fields` / `Item` / `Items` / `ValueOf` / `EmptyResult`

## 2. 业务标签当前协议

当前业务标签约定：

- 模板侧入口固定为品牌化业务指令，例如 `@sq:arc`
- Go 侧注册时使用业务名，例如 `blade.WithDataDirective("arc", handler)`
- 直接输出 HTML 的业务标签使用 `blade.WithTag("pager", handler)`，模板侧写 `@sq:pager(...)`
- 所有外部注册指令只接受业务局部名，统一映射到固定 `sq:` 入口；与内建流程指令冲突时会在初始化阶段失败
- 常规参数使用 `key=value`
- `out` 支持裸参数简写
- 裸 `out` 默认导出到 `field`
- `as=field`、`out=articles` 这类业务标签变量名参数会按变量名处理，不会读取同名数据变量
- 数据业务块只接受 `Items`，并直接逐项渲染块内容，无需嵌套 `@list`
- 数据业务块支持 `@empty` 空分支，并使用 `@end` 结束
- `out` 模式只负责把结果写入当前作用域；需要遍历导出的列表时再使用 `@list`
- 单对象或复合数据使用 `Value`
- 单对象空结果建议返回 `blade.EmptyResult()`
- 列表块模式空结果返回空 `Items`

## 3. 数据传递规则

模板组合的作用域规则：

- `@include('view')` 沿用当前作用域，局部模板可以读取根数据和当前循环变量。
- `@include('view', data)` 会把 `data` 展开为局部变量，并保留原有根数据与局部变量。
- `@component('view')` 会复制当前作用域，并额外注入 `slot` 和具名 slot。
- `@component('view', data)` 会把 `data` 展开为组件局部变量，并保留根数据、当前循环变量、默认 slot 和具名 slot。
- `@extends` 不支持传入第二参数；布局模板读取同一个根 `ViewContext`，页面差异通过 `@section` 和根上下文表达。

示例：

```blade
@include('partials.note', Page.IncludeData)

@component('components.card', Page.CardData)
  默认 slot 内容
@endcomponent
```

## 4. 文件模板能力

模板文件名当前支持：

- 不带后缀调用
- 点号路径，例如 `default.view`
- 命名空间路径，例如 `cms:default.view`
- 多层查找，例如 `blade.LayeredFS{Layers: []blade.TemplateFS{...}}`

默认可解析后缀：

- `.tpl`
- `.html`

`LayeredFS` 会按层顺序查找模板，适合主题覆盖、应用覆盖和模块默认模板。只有当前层返回未找到时才继续下一层，其他错误会直接返回。使用 `OSFS` 作为层时，`../`、绝对路径和符号链接逃逸都会被拒绝。

## 5. 错误提示

编译和解析错误当前会带：

- 错误消息
- 行列号
- 源码行
- 指针位置
- 文件模板场景下的模板名

## 6. 已知限制

- 表达式语言当前是受控子集，不等价于完整 Blade / PHP 表达式
- 自定义块指令当前只支持单层包裹式协议
- 业务标签当前仍是实验性 API
- 业务标签当前不支持对象参数、数组参数、查询串参数
- `@extends` 当前不支持独立传参，布局数据应来自根 `ViewContext`
- 组件当前只支持视图组件，不包含更完整的 props 绑定协议
- 默认 AST、编译文本和文件模板缓存分别最多保存 512 条；超出后按写入顺序淘汰
- 文件缓存当前还没有更高级的依赖图失效策略
