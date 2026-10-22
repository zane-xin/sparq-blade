# 业务标签指令设计

本文档记录 `v0.3.0` 计划中的业务标签指令方向，用于支持类似 CMS 标签、插件标签、数据块标签这类场景。

## 一、目标

业务标签指令要解决的问题：

- 模板中需要调用业务标签，例如文章、栏目、推荐位、插件数据源
- 标签通常有目标标识，例如 `cms:arc`
- 标签通常带命名参数，例如 `limit=1,pid=1`
- 部分标签可能需要块内容

第一版目标是提供稳定、可维护的保守方案，不直接扩展 Blade 指令名语法。

## 二、不采用的语法

暂不支持这种形式：

```blade
@sq:cms:arc(limit=1,pid=1)
  ...
@end:cms::arc
```

原因：

- 当前 lexer 的指令名只支持字母、数字和下划线
- 带冒号的原生命名空间指令名会扩大词法规则和结束指令匹配复杂度
- `@end:cms::arc` 这类闭合形式不够统一，错误恢复和提示成本高
- 该能力和模板命名空间 `cms:default.view` 不是同一层问题，不应混在一起

## 三、推荐语法

### 单指令标签

推荐：

```blade
@sq('cms:arc', limit=1, pid=1)
```

含义：

- `pt` 是普通 Blade 指令名
- `'cms:arc'` 是业务标签目标
- `limit=1,pid=1` 是命名参数

### 块标签

推荐：

```blade
@sq('cms:arc', limit=1, pid=1)
  ...
@endsq
```

含义：

- 同一个 `pt` 指令同时支持单指令和块指令
- 是否允许块内容由业务标签实现决定
- 结束指令固定为 `@endsq`

## 四、参数协议

第一版建议支持：

```blade
@sq('cms:arc', limit=1, pid=1, type='news', active=true)
@sq('cms:arc', limit=1, pid=1, out)
```

参数解析结果：

```go
TagCall{
    Target: "cms:arc",
    Params: map[string]any{
        "limit": 1,
        "pid": 1,
        "type": "news",
        "active": true,
    },
}
```

补充约定：

- 常规参数仍使用 `key=value`
- `out` 允许使用裸参数简写，默认导出到 `field`

### 支持的参数值

第一版建议支持：

- 字符串：`'news'`、`"news"`
- 数字：`1`、`1.5`
- 布尔：`true`、`false`
- `nil` / `null`
- 变量表达式：`user.id`、`category.id`

暂不支持：

- map / object 字面量
- array 字面量
- 嵌套函数调用
- 任意 Go template 语法

### 备选语法

如果后续需要更复杂参数，可再评估：

```blade
@sq('cms:arc', query)
@sq("cms:arc?limit=1&pid=1")
```

第一版不优先实现查询串形式，因为类型表达能力弱。

## 五、运行时协议

建议新增专门的数据标签 API，而不是让业务方手写 Go template 片段。

候选 API：

```go
type DataTagCall interface {
    Target() string
    Has(key string) bool
    GetString(key string) (string, error)
    GetStringDefault(key string, fallback string) (string, error)
    GetInt(key string) (int, error)
    GetIntDefault(key string, fallback int) (int, error)
    GetBool(key string) (bool, error)
    GetBoolDefault(key string, fallback bool) (bool, error)
    Alias() (string, error)
    Out() (string, bool, error)
}

type DataTagResult struct {
    Value any
    Items []any
}

type DataTagFunc func(call DataTagCall) (DataTagResult, error)

blade.WithDataTag("cms:arc", handler)
```

模板：

```blade
@sq('cms:arc', limit=1)
@sq('cms:list', limit=10)
  <li>{{ title }}</li>
@endsq
```

说明：

- 模板侧入口固定为 `pt`
- Go 侧按目标分别注册，例如 `cms:arc`
- `Target` 决定具体业务标签
- `Alias()` 对应 `id=field`
- `Out()` 对应 `out` 或 `out='arc'`
- `Value` 用于单对象导出
- `Items` 用于列表块渲染或列表导出
- 如需减少样板代码，可用 `blade.Fields()`、`blade.Item()`、`blade.Items()`、`blade.ValueOf()`

## 六、CMS 场景推荐约定

对于你当前这类 CMS 业务标签，建议统一采用下面的模板约定：

```blade
@sq("cms:arc", pid=10, limit=20)
  <li>{{ field.title }}</li>
@endsq

@sq("cms:arc", pid=10, limit=20, out)
@foreach(field as item)
  <li>{{ item.title }}</li>
@endforeach

@sq("cms:site", out)
<h1>{{ field.title }}</h1>

@sq("cms:arc", pid=10, limit=20, out=arc)
@foreach(arc as item)
  <li>{{ item.title }}</li>
@endforeach
```

推荐原因：

- 块模式默认当前项就是 `field`，和历史前端模板习惯一致
- 单指令裸 `out` 也默认导出到 `field`，减少两套记忆成本
- 只有在同一模板里需要同时保留多个数据结果时，再显式使用 `out=arc` 这类命名

## 七、推荐返回结构

为了避免每个业务标签都重新定义字段名，建议先约定一套 CMS 常用字段。

### 1. 列表型标签

例如 `cms:arc`、`cms:product`、`cms:banner` 这类“返回多条记录”的标签，建议返回：

```go
blade.DataTagResult{
    Items: []any{
        map[string]any{
            "id":      1,
            "title":   "Article A",
            "url":     "/article/a",
            "cover":   "/uploads/a.jpg",
            "summary": "Article summary",
            "time":    "2026-04-24",
        },
    },
}
```

建议字段：

- `id`: 主键或业务唯一标识
- `title`: 标题
- `url`: 前台访问地址
- `cover`: 封面图地址，没有时可置空
- `summary`: 摘要、描述或导语，没有时可置空
- `time`: 展示用时间字符串，没有时可置空

模板侧就可以保持稳定写法：

```blade
@sq("cms:arc", pid=10, limit=20)
  <li>
    <a href="{{ field.url }}">{{ field.title }}</a>
    @if(field.summary)
      <p>{{ field.summary }}</p>
    @endif
  </li>
@endsq
```

### 2. 单对象标签

例如 `cms:site`、`cms:category`、`cms:config` 这类“返回单个对象”的标签，建议返回：

```go
blade.DataTagResult{
    Value: map[string]any{
        "id":    1,
        "title": "SQ CMS",
        "url":   "https://example.test",
        "logo":  "/uploads/logo.png",
    },
}
```

建议字段：

- `id`: 对象标识，没有时可省略
- `title`: 对象名称
- `url`: 对象链接或主地址，没有时可置空
- `logo`: 站点 logo、栏目图标等扩展字段，没有时可置空

模板侧推荐：

```blade
@sq("cms:site", out)
<h1>{{ field.title }}</h1>
@if(field.logo)
  <img src="{{ field.logo }}" alt="{{ field.title }}">
@endif
```

### 3. 需要分页或复合数据时

如果后面遇到分页、列表加统计信息、列表加栏目对象这类“单纯 `Items` 不够用”的场景，建议不要把块模式做复杂，而是统一走单指令 `out`：

```go
blade.DataTagResult{
    Value: map[string]any{
        "items": items,
        "pager": map[string]any{
            "page":  1,
            "size":  20,
            "total": 135,
        },
        "category": category,
    },
}
```

模板侧：

```blade
@sq("cms:arc", pid=10, limit=20, out=page)
@foreach(page.items as item)
  <li>{{ item.title }}</li>
@endforeach
<span>{{ page.pager.total }}</span>
```

这套约定的好处是：

- 简单列表继续用块模式，模板最短
- 复合数据统一走 `out`，避免块模式协议膨胀
- 同类业务标签的字段名尽量稳定，前端模板迁移成本低

## 八、空结果约定

为了让模板层的判空语义稳定，建议统一采用下面的约定：

- 列表型标签在块模式下查不到数据时，直接返回空 `Items`，模板输出为空，不报错
- 列表型标签在单指令 `out` 模式下查不到数据时，可返回空 `Items` 或零值 `DataTagResult{}`，模板侧统一按空处理
- 单对象标签查不到数据时，建议直接返回 `DataTagResult{}`，模板侧 `field` 为 `nil`
- 不建议为了“空结果”构造字段全为空字符串的假对象，这会让模板判空变得不稳定

推荐模板写法：

```blade
@sq("cms:site", out)
@if(field)
  <h1>{{ field.title }}</h1>
@else
  <span>empty</span>
@endif

<ul>
@sq("cms:arc", pid=10)
  <li>{{ field.title }}</li>
@endsq
</ul>
```

这套约定的效果是：

- 单对象标签更适合用 `@if(field)` 判断是否存在
- 列表标签块模式天然“无数据即无输出”
- 复杂空态展示可以交给外围模板自行控制

## 九、和现有自定义指令的关系

现有能力：

- `blade.WithDirective`
- `blade.WithBlockDirective`

这些 API 直接返回 Go template 片段，适合底层扩展，但对业务标签不够安全。

数据标签 API 应作为更高层封装：

- 调用方返回数据结果，而不是 Go template 片段
- 不需要直接拼 Go template action
- 参数由引擎解析
- 引擎负责循环、变量注入和错误上下文

## 十、第一版实现建议

建议 `v0.3.0` 第一阶段只做：

1. 增加 `DataTagCall`、`DataTagFunc`、`WithDataTag`
2. 支持 `@sq('cms:arc', limit=1, pid=1)`
3. 支持 `@sq(...) ... @endsq`
4. 参数支持字符串、数字、布尔、nil、变量表达式
5. 支持 `id=field` 与 `out` / `out='arc'`
5. 增加单指令、块指令、错误路径测试
6. 更新 README、QUICKSTART、SUPPORT、CHANGELOG

暂不做：

- `@sq:cms:arc`
- `@end:cms::arc`
- 多阶段标签块
- 嵌套对象参数
- 查询串参数语法

## 十一、待确认问题

- 是否需要为 `loop` 增加更多字段
- 是否要支持更复杂的参数类型，例如对象和数组
- 是否需要补一个可运行的 `examples/tags`
