# Integration Checklist

本文档用于把当前 `blade` 的业务标签能力接入真实 CMS 项目前的检查项收成一份可执行清单。

## 1. 边界确认

先确认团队接受下面这些边界：

- 业务入口固定使用品牌指令，例如 `@sq:arc(...)`；流程指令使用 `@if`、`@list`、`@empty`、`@end` 等无品牌写法
- 指令前缀固定为 `@`，暂不支持自定义前缀
- 块模式默认当前项变量是 `field`，推荐显式使用 `as=article`
- 裸 `out` 默认也导出到 `field`
- 需要多个结果并存时显式写 `out=arc`、`out=page`
- 块模式只接受 `Items`
- 数据业务块直接遍历 `Items`，块内不再嵌套 `@list`，空结果通过 `@empty` 处理
- 单对象或复合数据统一走 `Value` + `out`
- 请求级数据统一从 `call.Context()` 获取，不使用包级全局变量保存请求状态

如果这些边界和项目现状冲突，建议先确认，不要边接入边改协议。

## 2. 模板检查

建议先拿 3 到 5 个真实模板验证。

### 列表块模式

```blade
@sq:arc(limit=20, as=article)
  <li><a href="{{ article.url }}">{{ article.title }}</a></li>
@end
```

### 显式栏目

```blade
@sq:arc(cid=10, limit=20, as=article)
  <li>{{ article.title }}</li>
@end
```

### 单对象导出

```blade
@sq:site(out)
@if(field)
  <h1>{{ field.title }}</h1>
@endif
```

### 复合分页数据

```blade
@sq:arc_page(cid=10, limit=20, out=page)
@list(page.items, as=item)
  <li>{{ item.title }}</li>
@empty
  <li>暂无内容</li>
@end
```

确认点：

- 是否接受显式 `as=article`
- 是否需要 `loop.index`、`loop.first`、`loop.last`
- 是否接受分页统一走 `Value`
- 是否存在必须迁移的历史模板写法

## 3. 请求上下文检查

推荐控制器为每次渲染传入一个明确的上下文对象：

```go
type ViewContext struct {
    Site  SiteContext
    Route RouteContext
    Page  PageData
    Meta  PageMeta
}
```

检查项：

- `@sq:title`、`@sq:seo` 是否只从当前 `ViewContext` 读取数据
- `@sq:arc()` 是否能从 `ViewContext.Page.CategoryID` 推导栏目
- `@sq:arc(cid=99)` 是否能覆盖上下文推导
- `Engine` 是否可以复用，`ViewContext` 是否按请求创建
- handler 中是否避免写入包级全局变量

推荐优先级：

```text
显式参数 > 页面配置 > 控制器写入 > 当前路径配置 > 全局系统配置
```

## 4. Handler 检查

每个业务标签 handler 建议都确认：

- 参数只通过 `GetIntDefault`、`GetStringDefault`、`GetBoolDefault` 读取
- 列表优先走 `Rows`
- 单对象优先走 `Object`
- 空结果统一走 `EmptyResult()` 或空 `Items`
- 不同时设置 `Value` 和 `Items`

## 5. 字段约定检查

接入前建议统一最小字段协议。

列表项建议：

- `id`
- `title`
- `url`
- `cover`
- `summary`
- `time`

单对象建议：

- `id`
- `title`
- `url`
- `logo`

分页对象建议：

- `items`
- `pager.page`
- `pager.size`
- `pager.total`

## 6. 命名空间与模板路径检查

如果项目已经按模块拆分模板，建议至少验证一次：

- `default.view`
- `cms:default.view`
- `@extends('cms:layout.default')`
- `@include('cms:partials.footer')`

## 7. 错误场景检查

建议故意验证这些错误是否容易定位：

- 未注册标签目标
- 参数类型错误
- 块模式误返回 `Value`
- `Value` 和 `Items` 同时设置
- 模板里引用不存在字段

## 8. 最小验收

建议至少跑通下面四类模板：

1. 首页文章列表
2. 栏目页分页列表
3. 站点配置或栏目对象读取
4. 空数据模板

通过标准：

- 模板能正常渲染
- 空结果行为符合预期
- 不需要继续改协议就能覆盖主场景
