@extends('layouts.default')

@section('content')
<section>
  <h2>直接子级渲染</h2>
  <p><code>&#64;children</code> 默认读取当前项的 <code>children</code> 字段，只渲染直接子级；更深层级需要在模板中继续嵌套。</p>
  <pre>@@sq:nav(as=item)
  &#123;&#123; item.title &#125;&#125;
  @@children(as=child, wrap="ul", class="submenu")
    &#123;&#123; child.title &#125;&#125;
  @@empty
    无子级
  @@end
@@end</pre>

  <ul class="demo-list">
  @sq:nav(as=item)
    <li>
      <a href="{{ item.url }}">{{ item.title }}</a>
      @children(as=child, wrap="ul", class="submenu")
        <li><a href="{{ child.url }}">{{ child.title }}</a></li>
      @empty
        <span>无子级</span>
      @end
    </li>
  @end
  </ul>
</section>
@endsection
