@extends('layout')

@section('content')
<section>
  <h2>{{ Page.Title }}</h2>
  <p>{{ Page.Summary }}</p>
  <p><span class="badge">模板指令</span> 本页使用 <code>&#64;component</code>、<code>&#64;slot</code> 和 <code>&#64;include("partials.steps")</code>。</p>
</section>

@include('partials.steps')

<section>
  <h2>组件效果</h2>
  @component('components.doc-card', Page.CardData)
    组件标题来自传入数据，组件正文来自默认 slot。真实项目中可以用它封装卡片、面板、表格行等重复 UI。
  @endcomponent
</section>

<section>
  <h2>模板写法</h2>
  <pre>{{ Page.TemplateCode }}</pre>
  <h2>Go 接入</h2>
  <pre>{{ Page.GoCode }}</pre>
</section>
@endsection
