@extends('layouts.default')

@section('content')
<section>
  <h2>布局继承</h2>
  <p>页面模板使用 <code>&#64;extends</code> 选择布局，用 <code>&#64;section</code> 填充布局中的 <code>&#64;yield</code>。</p>
  <pre>@@extends('layouts.default')

@@section('content')
  页面正文
@@end</pre>
  <p class="note">布局需要的站点、导航、面包屑等数据来自根 <code>ViewContext</code>，不在 <code>&#64;extends</code> 上单独传参。</p>
</section>

<section id="include">
  <h2>局部模板</h2>
  <p><code>&#64;include</code> 可以直接沿用当前作用域，也可以通过第二参数传入局部数据。</p>
  <pre>@@include('partials.doc-note', Page.Include)</pre>
  @include('partials.doc-note', Page.Include)
</section>

<section id="component">
  <h2>组件与插槽</h2>
  <p><code>&#64;component</code> 适合封装可复用模板结构；默认 slot 是组件正文，具名 slot 通过 <code>&#64;slot</code> 传入。</p>
  <pre>@@component('components.doc-card', Page.Card)
  @@slot('footer')
    来自 footer slot
  @@end
  组件正文
@@end</pre>

  @component('components.doc-card', Page.Card)
    @slot('footer')
      来自 footer slot 的补充信息。
    @end
    组件正文来自默认 slot，组件标题来自第二参数。
  @end
</section>

<section>
  <h2>each 渲染列表局部</h2>
  <p><code>&#64;each</code> 会为列表中的每一项渲染行模板，也可以提供空模板。</p>
  <pre>@@each('partials.doc-row', Page.Items, 'item', 'partials.empty')</pre>
  <div>
    @each('partials.doc-row', Page.Items, 'item', 'partials.empty')
  </div>
</section>

<section>
  <h2>命名空间模板</h2>
  <p>模板名可以使用 <code>cms:partials.pager</code> 这样的命名空间。当前示例通过 <code>LayeredFS</code> 让主题目录优先于模块目录。</p>
  <pre>blade.OSFS{
  Root: "themes/default/views",
  Namespaces: map[string]string{
    "cms": "themes/default/views/cms",
  },
}</pre>
</section>
@endsection
