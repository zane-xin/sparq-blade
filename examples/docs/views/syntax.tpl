@extends('layout')

@section('content')
<section>
  <h2>{{ Page.Title }}</h2>
  <p>{{ Page.Summary }}</p>
  <p><span class="badge">语法配置</span> 默认语法仍然是 <code>&#123;&#123; &#125;&#125;</code>、<code>&#123;!! !!&#125;</code>、<code>&#123;&#123;-- --&#125;&#125;</code>。如果项目需要避开前端框架语法冲突，可以在创建 Engine 时替换这些分隔符。</p>
</section>

@include('partials.steps')

<section>
  <h2>推荐配置方式</h2>
  <pre>{{ Page.GoCode }}</pre>
  <p class="muted">配置作用在当前 Engine 上。指令语法仍然保持 <code>&#64;if</code>、<code>&#64;list</code>、<code>&#64;sq:arc</code>、<code>&#64;sq:title</code> 这样的写法。</p>
</section>

<section>
  <h2>模板写法</h2>
  <pre>{{ Page.TemplateCode }}</pre>
  <ul>
    <li><code>[[ title ]]</code>：转义输出，等价于默认的 <code>&#123;&#123; title &#125;&#125;</code>。</li>
    <li><code>[[! html !]]</code>：原样输出，等价于默认的 <code>&#123;!! html !!&#125;</code>。</li>
    <li><code>[[-- comment --]]</code>：模板注释，等价于默认的 <code>&#123;&#123;-- comment --&#125;&#125;</code>。</li>
  </ul>
</section>

<section>
  <h2>适用场景</h2>
  <ul>
    <li>模板需要和 Vue、Angular、Alpine 等前端模板共存。</li>
    <li>团队希望统一使用内部模板规范，但仍复用 Blade 的指令体系。</li>
    <li>已有 CMS 模板历史包袱，需要逐步迁移到 Go 渲染组件。</li>
  </ul>
</section>
@endsection
