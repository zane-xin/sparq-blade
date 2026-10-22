@extends('layout')

@section('content')
<section>
  <h2>{{ Page.Title }}</h2>
  <p>{{ Page.Summary }}</p>
  <p><span class="badge">模板指令</span> 本页使用 <code>&#64;sq:arc(cid=3, limit=2) ... &#64;end</code>。</p>
</section>

@include('partials.steps')

<section>
  <h2>文章列表</h2>
  <ul>
  @sq:arc(cid=3, limit=2, as=article)
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
