@extends('layout')

@section('content')
<section>
  <h2>{{ Page.Title }}</h2>
  <p>{{ Page.Summary }}</p>
  <p><span class="badge">模板指令</span> 本页使用 <code>&#64;sq:arc(...)</code>，并通过 <code>call.Context()</code> 自动推导栏目。</p>
</section>

@include('partials.steps')

<section>
  <h2>上下文推导列表</h2>
  <p class="muted">当前栏目：{{ Page.CategoryName }}，栏目 ID：{{ Page.CategoryID }}</p>
  <ul>
  @sq:arc(limit=2, as=article)
    <li><a href="{{ article.url }}">{{ article.title }}</a></li>
  @end
  </ul>
</section>

<section>
  <h2>显式参数覆盖上下文</h2>
  <ul>
  @sq:arc(cid=99, limit=1, as=article)
    <li><a href="{{ article.url }}">{{ article.title }}</a></li>
  @end
  </ul>
</section>

<section>
  <h2>模板写法</h2>
  <pre>{{ Page.TemplateCode }}</pre>
  <h2>Go 接入</h2>
  <pre>{{ Page.GoCode }}</pre>
</section>
@endsection
