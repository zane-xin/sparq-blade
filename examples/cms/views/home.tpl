@extends('layouts.default')

@section('content')
<section>
  <h2>这个示例覆盖什么</h2>
  <ul class="demo-list">
  @list(Docs, as=item)
    <li><strong>{{ item.Title }}</strong>: {{ item.Body }}</li>
  @empty
    <li>暂无说明。</li>
  @end
  </ul>
</section>

<section>
  <h2>路由说明</h2>
  <ul class="demo-list">
    <li><a href="/directives">/directives</a>：基础输出、条件、列表、空状态、计数循环和 <code>&#64;go</code>。</li>
    <li><a href="/templates">/templates</a>：布局继承、include、component、slot、each 和命名空间模板。</li>
    <li><a href="/data">/data</a>：数据指令返回列表、单对象、复合对象和空结果。</li>
    <li><a href="/articles">/articles</a>：数据指令、上下文推导和显式参数覆盖。</li>
    <li><a href="/navigation">/navigation</a>：使用 <code>&#64;children</code> 渲染当前项的直接子级。</li>
    <li><a href="/pagination">/pagination</a>：使用 <code>&#64;sq:pager</code> 输出分页，并验证 LayeredFS 主题覆盖。</li>
    <li><a href="/custom">/custom</a>：自定义单指令、块级指令、输出型业务标签和数据型业务指令。</li>
  </ul>
</section>

<section>
  <h2>模板查找层级</h2>
  <pre>blade.LayeredFS{
  Layers: []blade.TemplateFS{
    blade.OSFS{Root: "themes/default/views"},
    blade.OSFS{Root: "views"},
    blade.OSFS{Root: "modules/cms/views"},
  },
}</pre>
  <p class="note">模板里保持稳定名称，Go 侧通过 FS 层级决定当前主题、应用模板和模块默认模板的优先级。</p>
</section>
@endsection
