@extends('layouts.default')

@section('content')
<section>
  <h2>输出型业务标签</h2>
  <p><code>&#64;sq:pager()</code> 由 <code>WithTag("pager")</code> 注册。handler 可以从上下文推导分页数据，并通过 <code>call.Render</code> 渲染局部模板。</p>
  <pre>@@sq:pager(class="pager pager-theme")</pre>

  @sq:pager(class="pager pager-theme")
</section>

<section>
  <h2>LayeredFS 主题覆盖</h2>
  <p>handler 渲染 <code>cms:partials.pager</code>。当前主题中存在同名命名空间模板，因此会覆盖模块中的默认分页模板。</p>
  <ul class="demo-list">
    <li>主题模板：<code>themes/default/views/cms/partials/pager.tpl</code></li>
    <li>模块默认：<code>modules/cms/views/partials/pager.tpl</code></li>
  </ul>
</section>
@endsection
