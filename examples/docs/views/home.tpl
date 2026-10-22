@extends('layout')

@section('content')
<section>
  <h2>应用场景入口</h2>
  <p>这个示例站点把 Blade 的常见能力组织成可以直接访问的操作文档。每个页面都对应一种接入场景，并在模板中真实使用相关指令。</p>
  <div class="grid">
  @sq:scenarios(as=item)
    @component('components.doc-card')
      @slot('title')
        <a href="{{ item.url }}">{{ item.title }}</a>
      @endslot
      {{ item.description }}
    @endcomponent
  @end
  </div>
</section>

<section>
  <h2>推荐顺序</h2>
  <ol>
    <li>先看基础渲染，理解 Engine、模板目录和 RenderFile。</li>
    <li>再看组件，理解如何复用页面片段。</li>
    <li>再看数据标签，理解业务数据如何进入模板。</li>
    <li>最后看请求上下文和 HTTP 集成，理解真实项目中的组织方式。</li>
  </ol>
</section>

<section>
  <h2>渲染流程</h2>
  <pre class="mermaid">flowchart LR
  A["HTTP 请求"] --> B["控制器构造 ViewContext"]
  B --> C["Engine.RenderFile"]
  C --> D["解析模板指令"]
  D --> E["@@sq:title / @@sq:seo"]
  D --> F["@@sq 数据标签"]
  E --> G["call.Context() 读取上下文"]
  F --> G
  G --> H["输出 HTML"]</pre>
</section>
@endsection
