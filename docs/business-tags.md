# 业务标签

本文档集中说明当前 `@sq` 业务标签的推荐用法、返回结构和接入约定。

## 1. 设计边界

当前固定采用：

```blade
@sq:arc(pid=10, limit=20, as=article)
  ...
@end
```

业务侧通过 `blade.WithDataDirective("arc", handler)` 注册业务指令，模板中固定使用品牌化入口 `@sq:arc`。

数据业务块使用流程指令 `@end` 结束，并可通过 `@empty` 描述空结果。`sq:` 只出现在业务指令名上，流程指令不使用品牌前缀。

## 2. 模板写法

### 列表块模式

```blade
@sq:arc(pid=10, limit=20, as=article)
  <li><a href="{{ article.url }}">{{ article.title }}</a></li>
@empty
  <li>暂无内容</li>
@end
```

说明：

- 默认当前项变量是 `field`，推荐通过 `as=article` 显式命名
- 会注入 `loop.index`、`loop.first`、`loop.last` 等循环字段
- 业务块会直接遍历 handler 返回的 `Items`，块内不需要再嵌套 `@list`
- `Items` 为空时渲染 `@empty` 分支；未声明 `@empty` 时不输出内容

### 上下文自动推导

当前访问路径已经能推导栏目时，模板可以不显式传 `cid`：

```blade
@sq:arc(limit=10, as=article)
  <li>{{ article.title }}</li>
@end
```

Handler 中推荐这样处理：

```go
cid, err := call.GetIntDefault("cid", 0)
if err != nil {
    return blade.DataTagResult{}, err
}
if cid == 0 {
    ctx := call.Context().(ViewContext)
    cid = ctx.Page.CategoryID
}
```

优先级约定：

```text
显式参数 > 当前页面上下文 > 当前路径配置 > 全局配置
```

因此：

- `@sq:arc()` 使用当前访问上下文推导栏目
- `@sq:arc(cid=1)` 使用显式栏目 ID
- `@sq:arc(out=articles)` 导出列表到 `articles`

### 层级子级

当当前项包含 `children` 字段时，可以用 `@children` 渲染直接子级。该指令不自动递归，需要更深层级时在模板中继续嵌套一层。

```blade
@sq:nav(as=item)
  <li>
    <a href="{{ item.url }}">{{ item.title }}</a>

    @children(as=child, wrap="ul", class="submenu")
      <li>{{ child.title }}</li>
    @empty
      <li>无嵌套</li>
    @end
  </li>
@end
```

如果子级字段不是 `children`，可以显式指定：

```blade
@children(field=nodes, as=node, wrap="div", class="children")
  <section>{{ node.title }}</section>
@end
```

### 单对象导出

```blade
@sq:site(out)
@if(field)
  <h1>{{ field.title }}</h1>
@endif
```

说明：

- 裸 `out` 默认导出到 `field`
- 单对象建议通过 `@if(field)` 判空

### 多结果并存

```blade
@sq:arc(pid=10, limit=20, out=arc)
@list(arc, as=item)
  <li>{{ item.title }}</li>
@end
```

说明：

- 只有需要多个结果并存时，再显式命名 `out=arc`
- `out` 模式不会自动遍历结果，因此导出后使用流程指令 `@list` 渲染列表

### 复合分页数据

```blade
@sq:arc_page(pid=10, limit=20, out=page)
@list(page.items, as=item)
  <li>{{ item.title }}</li>
@end
<span>{{ page.pager.total }}</span>
```

### 输出型标签

分页、面包屑这类业务可以直接输出 HTML，避免模板里重复拼接结构：

```blade
@sq:pager()
@sq:pager(template="partials.pager", class="pager")
@sq:crumb()
```

推荐约定：

- `@sq:pager()` 默认从当前渲染上下文推导分页对象
- `for=...` 只在需要覆盖默认推导时传入
- `template=...` 指定业务侧可维护的局部模板
- `class=...` 等参数只作为渲染配置，不改变数据来源优先级

Go 侧使用 `WithTag` 注册：

```go
renderer := blade.New(
    blade.WithTag("pager", func(call blade.TagCall) (template.HTML, error) {
        pager, ok := call.Get("for")
        if !ok {
            ctx := call.Context().(ViewContext)
            pager = ctx.Page.Pager
        }

        tpl, err := call.GetStringDefault("template", "")
        if err != nil {
            return "", err
        }
        className, err := call.GetStringDefault("class", "pager")
        if err != nil {
            return "", err
        }

        data := map[string]any{
            "pager": pager,
            "class": className,
        }

        if tpl != "" {
            return call.Render(tpl, data)
        }

        return renderDefaultPager(data), nil
    }),
)
```

模板文件示例：

```blade
<nav class="{{ class }}">
  <span>{{ pager.page }}/{{ pager.total }}</span>
</nav>
```

## 3. Go 侧注册

```go
renderer := blade.New(
    blade.WithDataDirective("arc", handler),
    blade.WithDataDirective("site", handler),
)
```

参数读取推荐统一使用：

```go
pid, err := call.GetIntDefault("pid", 0)
limit, err := call.GetIntDefault("limit", 10)
key, err := call.GetStringDefault("key", "default")
enabled, err := call.GetBoolDefault("enabled", false)
```

避免零散 map 断言和重复类型转换。

## 4. 返回结构

### 列表结果

```go
return blade.Rows(
    map[string]any{
        "id":    1,
        "title": "Article A",
        "url":   "/a",
    },
), nil
```

`Rows` 也可以直接接收结构体或结构体切片：

```go
type Article struct {
    ID    int
    Title string
    URL   string
}

return blade.Rows([]Article{
    {ID: 1, Title: "Article A", URL: "/a"},
    {ID: 2, Title: "Article B", URL: "/b"},
}), nil
```

### 单对象结果

```go
return blade.Object(map[string]any{
    "title": "SQ CMS",
    "url":   "https://example.test",
}), nil
```

`Object` 同样支持结构体。

### 空结果

```go
return blade.EmptyResult(), nil
```

空结果约定：

- 单对象无数据：返回 `blade.EmptyResult()`
- 列表块模式无数据：返回空 `Items`
- 不建议构造“字段全为空字符串”的假对象

## 5. 推荐字段约定

### 列表型标签

建议字段：

- `id`
- `title`
- `url`
- `cover`
- `summary`
- `time`

### 单对象标签

建议字段：

- `id`
- `title`
- `url`
- `logo`

### 分页对象

建议字段：

- `items`
- `pager.page`
- `pager.size`
- `pager.total`

## 6. Helper API

当前已提供轻量 helper，减少 handler 样板代码：

- `blade.Rows(...)`
- `blade.Object(...)`
- `blade.EmptyResult()`

需要链式构造字段时，也可以继续使用：

- `blade.Fields()`
- `blade.Item(...)`
- `blade.Items(...)`
- `blade.ValueOf(...)`

这层 helper 的目标是减少重复，不是引入新的 DSL。

## 7. 推荐接入顺序

1. 先接 `arc` 这类最简单的列表数据指令，模板侧使用 `@sq:arc`
2. 再接 `site` 这类单对象数据指令，模板侧使用 `@sq:site`
3. 再接 `arc_page` 这类复合分页数据指令，模板侧使用 `@sq:arc_page`
4. 最后再处理历史模板里最复杂的场景

如果准备进入真实项目接入，继续看 [Integration Checklist](integration-checklist.md)。
