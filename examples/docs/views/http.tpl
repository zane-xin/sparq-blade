@extends('layout')

@section('content')
<section>
  <h2>{{ Page.Title }}</h2>
  <p>{{ Page.Summary }}</p>
  <p><span class="badge">运行地址</span> `http://127.0.0.1:8082/`</p>
</section>

@include('partials.steps')

<section>
  <h2>HTTP 接入要点</h2>
  <ul>
    <li>应用启动时创建并复用一个 `Engine`。</li>
    <li>每个请求构造自己的 `ViewContext`。</li>
    <li>业务标签通过 `call.Context()` 读取当前请求数据。</li>
    <li>不要把请求数据写入包级全局变量。</li>
  </ul>
</section>

<section>
  <h2>模板写法</h2>
  <pre>{{ Page.TemplateCode }}</pre>
  <h2>Go 接入</h2>
  <pre>{{ Page.GoCode }}</pre>
</section>
@endsection
