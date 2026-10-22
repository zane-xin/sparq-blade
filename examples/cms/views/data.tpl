@extends('layouts.default')

@section('content')
<section>
  <h2>列表结果：Items / Rows</h2>
  <p>数据型业务指令返回列表时，可以直接作为块指令使用，并用 <code>as</code> 指定当前项别名。</p>
  <pre>@@sq:arc(limit=2, as=article)
  &#123;&#123; article.title &#125;&#125;
@@end</pre>
  <ul class="demo-list">
  @sq:arc(limit=2, as=article)
    <li><a href="{{ article.url }}">{{ article.title }}</a> <span class="meta">{{ article.time }}</span></li>
  @end
  </ul>
</section>

<section>
  <h2>单对象：ValueOf / Object</h2>
  <p>单对象更适合用 <code>out</code> 输出到变量，再用条件判断做空值处理。</p>
  <pre>@@sq:site(out)
@@if(field)
  &#123;&#123; field.name &#125;&#125;
@@end</pre>

  @sq:site(out)
  @if(field)
    <p><strong>{{ field.name }}</strong>：{{ field.url }}，{{ field.icp }}</p>
  @else
    <p>站点信息为空。</p>
  @end
</section>

<section>
  <h2>复合页面数据</h2>
  <p>分页列表这类场景可以返回一个对象，包含 <code>items</code>、<code>pager</code> 和栏目数据。分页标签默认能从当前上下文推导，也可以显式传入 <code>for</code>。</p>
  <pre>@@sq:arc_page(cid=7, limit=2, out=page)
@@list(page.items, as=item)
  &#123;&#123; item.title &#125;&#125;
@@end
@@sq:pager(for=page.pager)</pre>

  @sq:arc_page(cid=7, limit=2, out=page)
  <h3>{{ page.category.title }}</h3>
  <ul class="demo-list">
  @list(page.items, as=item)
    <li><a href="{{ item.url }}">{{ item.title }}</a></li>
  @empty
    <li>暂无文章</li>
  @end
  </ul>
  @sq:pager(for=page.pager, class="pager pager-theme")
</section>

<section>
  <h2>空结果：EmptyResult</h2>
  <p>handler 查不到单对象时返回 <code>blade.EmptyResult()</code>，模板侧用 <code>&#64;if(field)</code> 判断。</p>
  <pre>@@sq:site(key='missing', out)
@@if(field)
  有数据
@@else
  站点信息为空
@@end</pre>

  @sq:site(key='missing', out)
  @if(field)
    <p>{{ field.name }}</p>
  @else
    <p class="note">站点信息为空。</p>
  @end
</section>
@endsection
