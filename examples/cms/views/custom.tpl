@extends('layouts.default')

@section('content')
<section>
  <h2>注册规则</h2>
  <p>Go 侧注册业务名，引擎根据品牌前缀提供模板入口。内置指令、自定义指令和业务标签不能重名。</p>
  <table>
    <thead>
      <tr><th>Go API</th><th>模板侧</th><th>适用场景</th></tr>
    </thead>
    <tbody>
      <tr><td><code>WithDirective("badge")</code></td><td><code>&#64;sq:badge(...)</code></td><td>编译期替换单个指令。</td></tr>
      <tr><td><code>WithBlockDirective("panel")</code></td><td><code>&#64;sq:panel ... &#64;end</code></td><td>编译期包裹一段模板内容。</td></tr>
      <tr><td><code>WithTag("pager")</code></td><td><code>&#64;sq:pager(...)</code></td><td>运行期输出 HTML 片段。</td></tr>
      <tr><td><code>WithDataDirective("arc")</code></td><td><code>&#64;sq:arc ... &#64;end</code></td><td>运行期取数据并驱动块渲染。</td></tr>
    </tbody>
  </table>
</section>

<section>
  <h2>单指令</h2>
  <p><code>&#64;sq:badge</code> 在 Go 侧编译为一个带样式的安全输出片段。</p>
  <pre>blade.WithDirective("badge", func(expression string) (string, error) {
    return fmt.Sprintf(`&lt;span class="badge"&gt;&#123;&#123; bladeEcho . %q &#125;&#125;&lt;/span&gt;`, expression), nil
})

@@sq:badge(Page.Title)</pre>
  <p>@sq:badge(Page.Title)</p>
</section>

<section>
  <h2>块级指令</h2>
  <p><code>&#64;sq:panel</code> 适合做轻量模板宏；块内部仍会继续按模板语法渲染。</p>
  <pre>@@sq:panel(document)
  &#123;&#123; Page.Summary &#125;&#125;
@@end</pre>

  @sq:panel(document)
    <p>{{ Page.Summary }}</p>
  @end
</section>

<section>
  <h2>输出型业务标签</h2>
  <p><code>WithTag</code> 适合分页、面包屑、SEO 片段这类需要拼装 HTML 或渲染局部模板的能力。</p>
  <pre>blade.WithTag("pager", renderPager)

@@sq:pager()</pre>
  @sq:pager()
</section>

<section>
  <h2>数据型业务指令</h2>
  <p><code>WithDataDirective</code> 负责把业务数据转成模板可使用的列表或对象。模板侧固定使用品牌化入口，避免业务标签和基础指令混在一起。</p>
  <pre>blade.WithDataDirective("arc", renderArticles)

@@sq:arc(limit=1, as=article)
  &#123;&#123; article.title &#125;&#125;
@@end</pre>
  <ul class="demo-list">
  @sq:arc(limit=1, as=article)
    <li>{{ article.title }}</li>
  @end
  </ul>
</section>
@endsection
