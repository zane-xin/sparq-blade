@extends('layout')

@section('content')
<h1>{{ Page.Category.Name }}</h1>
<p>{{ Meta.Description }}</p>

<h2>从当前访问上下文自动推导栏目</h2>
<ul>
@sq:arc(limit=2, as=article)
  <li><a href="{{ article.url }}">{{ article.title }}</a></li>
@end
</ul>

<h2>模板参数显式覆盖栏目</h2>
<ul>
@sq:arc(cid=99, limit=1, as=article)
  <li><a href="{{ article.url }}">{{ article.title }}</a></li>
@end
</ul>

@endsection
