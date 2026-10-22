# CMS 接入检查清单

本文档用于把当前 `blade` 的业务标签能力接入到真实 CMS 项目前的验证项收成一份可执行清单。

适用场景：

- 已决定使用固定入口 `@sq(...)`
- 已准备通过 `blade.WithDataTag("cms:xxx", handler)` 注册业务标签
- 准备把现有 CMS 模板或标签系统迁移到当前实现

## 一、先确认的边界

接入前先确认团队已经接受下面这些边界：

- 业务标签入口固定为 `@sq(...)` / `@endsq`
- 块模式默认当前项变量是 `field`
- 裸 `out` 默认也导出到 `field`
- 需要多个结果并存时显式写 `out=arc`、`out=page`
- 块模式只接受 `Items`
- 单对象或复合数据统一走 `Value` + `out`
- 当前不支持 `@sq:cms:arc(...)` 这类带冒号的原生命名空间指令名

如果这些边界里有任何一条和项目现状冲突，建议先停下来确认，不要边接入边改协议。

## 二、模板语法检查

建议先挑 3 到 5 个真实模板，逐项确认下面这些写法是否够用：

### 1. 列表块模式

```blade
@sq("cms:arc", pid=10, limit=20)
  <li><a href="{{ field.url }}">{{ field.title }}</a></li>
@endsq
```

确认点：

- 是否习惯使用 `field`
- 是否需要 `loop.index`、`loop.first`、`loop.last`
- 是否存在旧模板依赖其他默认变量名

### 2. 单对象导出

```blade
@sq("cms:site", out)
@if(field)
  <h1>{{ field.title }}</h1>
@endif
```

确认点：

- 模板层是否接受裸 `out` 默认就是 `field`
- 单对象判空是否统一写成 `@if(field)`

### 3. 复合分页数据

```blade
@sq("cms:arc_page", pid=10, limit=20, out=page)
@foreach(page.items as item)
  <li>{{ item.title }}</li>
@endforeach
<span>{{ page.pager.total }}</span>
```

确认点：

- 分页、分类、统计信息是否都能通过 `Value` 统一承载
- 是否还需要额外的分页字段命名约定

## 三、Handler 协议检查

每个业务标签 handler 建议都检查下面这些点：

### 1. 参数读取

确认已经只使用统一读取方式：

```go
pid, err := call.GetIntDefault("pid", 0)
limit, err := call.GetIntDefault("limit", 10)
key, err := call.GetStringDefault("key", "default")
enabled, err := call.GetBoolDefault("enabled", false)
```

不要混用零散的 map 断言和字符串转型逻辑。

### 2. 返回结构

列表标签建议统一返回：

```go
return blade.Items(
    blade.Item(
        blade.Fields().
            Set("id", 1).
            Set("title", "Article A").
            Set("url", "/a"),
    ),
), nil
```

单对象标签建议统一返回：

```go
return blade.ValueOf(
    blade.Fields().
        Set("title", "SQ CMS").
        Set("url", "https://example.test"),
), nil
```

### 3. 空结果

统一按下面的规则处理：

- 列表块模式无数据：返回空 `Items`
- 单对象无数据：返回 `blade.EmptyResult()`
- 不要返回“字段全为空字符串”的假对象

## 四、字段约定检查

接入前建议先在业务层统一最小字段协议，至少确认下面这组命名：

### 1. 列表项

- `id`
- `title`
- `url`
- `cover`
- `summary`
- `time`

### 2. 单对象

- `id`
- `title`
- `url`
- `logo`

### 3. 分页对象

- `items`
- `pager.page`
- `pager.size`
- `pager.total`

如果现有项目字段名差异很大，建议在 handler 里先做一次映射，不要把旧字段名直接泄漏到模板层。

## 五、文件与命名空间检查

如果项目里同时有基础模板和模块模板，建议至少验证一次：

- `default.view`
- `cms:default.view`
- `@extends('cms:layout.default')`
- `@include('cms:partials.footer')`

确认点：

- 视图根目录是否已经拆对
- 模块命名空间是否和现有项目目录一致
- 旧模板名迁移后是否存在冲突

## 六、错误与回退检查

建议故意制造几类错误，确认报错是否足够定位问题：

- 未注册标签目标，例如 `cms:unknown`
- 参数类型错误，例如 `pid='abc'`
- 块模式误返回 `Value`
- `Value` 和 `Items` 同时设置
- 模板里引用不存在字段

如果这些错误已经能让开发者快速定位，就不用继续抽象额外机制。

## 七、接入前最小验收

在真实项目里，建议至少挑下面四类模板各跑一遍：

1. 首页文章列表
2. 栏目页分页列表
3. 站点配置或栏目对象读取
4. 空数据模板

通过标准：

- 模板能正常渲染
- 空结果行为符合预期
- 不需要额外改协议就能覆盖主场景
- 模板作者能看懂 `@sq` 约定并接受 `field` / `out`

## 八、建议接入顺序

推荐按下面顺序推进：

1. 先接一个最简单的列表标签，例如 `cms:arc`
2. 再接一个单对象标签，例如 `cms:site`
3. 再接一个复合分页标签，例如 `cms:arc_page`
4. 最后再处理历史模板中最复杂的场景

这样做的目的是先验证协议本身，再处理历史包袱。
