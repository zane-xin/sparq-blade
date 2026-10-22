@extends('layout')

@section('content')
<section>
  <h2>{{ Page.Title }}</h2>
  <p>{{ Page.Summary }}</p>
  <p><span class="badge">模板指令</span> 本页使用 <code>&#64;extends</code>、<code>&#64;section</code>、<code>&#64;yield</code>、<code>&#64;include</code>、<code>&#64;list</code>。</p>
</section>

@include('partials.steps')

@include('partials.doc-note', Page.IncludeData)

<section>
  <h2>模板写法</h2>
  <pre>{{ Page.TemplateCode }}</pre>
  <ul>
  @list(Page.Items, as=item)
    <li>{{ item.name }}</li>
  @end
  </ul>
</section>

<section>
  <h2>Go 接入</h2>
  <pre>{{ Page.GoCode }}</pre>
</section>
@endsection
