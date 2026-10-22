@extends('layouts.default')

@section('content')
<section>
  <h2>1. 基于上下文推导数据</h2>
  <p>当模板没有传入 <code>cid</code> 时，handler 从 <code>call.Context()</code> 读取当前栏目。</p>
  <pre>@@sq:arc(limit=2, as=article)
  ...
@@end</pre>

  <ul class="demo-list">
  @sq:arc(limit=2, as=article)
    <li>
      <a href="{{ article.url }}">{{ article.title }}</a>
      <small>{{ article.time }}</small>
      <p>{{ article.summary }}</p>
    </li>
  @end
  </ul>
</section>

<section>
  <h2>2. 显式参数覆盖上下文</h2>
  <p>传入 <code>cid=99</code> 后，业务 handler 优先使用模板参数，而不是当前页面上下文。</p>
  <pre>@@sq:arc(cid=99, limit=1, as=article)
  ...
@@end</pre>

  <ul class="demo-list">
  @sq:arc(cid=99, limit=1, as=article)
    <li><a href="{{ article.url }}">{{ article.title }}</a></li>
  @end
  </ul>
</section>
@endsection
