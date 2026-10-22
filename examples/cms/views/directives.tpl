@extends('layouts.default')

@section('content')
<section>
  <h2>输出与注释</h2>
  <p>普通输出会自动转义，原样输出只适合可信 HTML。模板注释不会进入最终 HTML。</p>
  <table>
    <thead>
      <tr><th>写法</th><th>用途</th></tr>
    </thead>
    <tbody>
      <tr><td><code>&#123;&#123; Page.Title &#125;&#125;</code></td><td>转义输出，默认推荐。</td></tr>
      <tr><td><code>&#123;!! html !!&#125;</code></td><td>原样输出，handler 或数据层需要保证安全。</td></tr>
      <tr><td><code>&#123;&#123;-- 注释 --&#125;&#125;</code></td><td>模板注释，不会输出。</td></tr>
    </tbody>
  </table>
  <p>当前页面标题：<strong>{{ Page.Title }}</strong></p>
</section>

<section>
  <h2>条件判断</h2>
  <pre>@@if(Page.CategoryID == 7)
  当前栏目是 7
@@elseif(Page.CategoryID)
  已设置栏目
@@else
  未设置栏目
@@end

@@unless(Page.CategoryID == 0)
  有可用栏目上下文
@@end</pre>

  @if(Page.CategoryID == 7)
    <p class="note">当前栏目是 7。</p>
  @elseif(Page.CategoryID)
    <p class="note">已设置栏目。</p>
  @else
    <p class="note">未设置栏目。</p>
  @end

  @unless(Page.CategoryID == 0)
    <p>unless 分支确认：有可用栏目上下文。</p>
  @end
</section>

<section>
  <h2>列表与空状态</h2>
  <p>普通列表推荐使用品牌化 <code>&#64;list</code>，并用 <code>&#64;empty</code> 描述空列表内容。</p>
  <pre>@@list(Page.Items, as=item)
  &#123;&#123; item.title &#125;&#125;
@@empty
  暂无内容
@@end</pre>

  <ul class="demo-list">
  @list(Page.Items, as=item)
    <li><strong>{{ item.title }}</strong> <span class="meta">{{ item.status }}</span></li>
  @empty
    <li>暂无内容</li>
  @end
  </ul>

  <p>空列表示例：</p>
  <ul class="demo-list">
  @list(UnknownItems, as=item)
    <li>{{ item.title }}</li>
  @empty
    <li>暂无内容</li>
  @end
  </ul>
</section>

<section>
  <h2>计数循环与高级遍历</h2>
  <p><code>&#64;for(i in 1..5)</code> 用于简单计数；<code>&#64;foreach</code> 仍可用，但更适合高阶场景。</p>
  <pre>@@for(i in 1..5)
  &#123;&#123; i &#125;&#125;
@@end

@@foreach(Page.Items as item)
  &#123;&#123; loop.index &#125;&#125;: &#123;&#123; item.title &#125;&#125;
@@end</pre>

  <p>
  @for(i in 1..5)
    <span class="badge">{{ i }}</span>
  @end
  </p>

  <ol>
  @foreach(Page.Items as item)
    <li>{{ loop.index }} / {{ loop.count }}：{{ item.title }}</li>
  @end
  </ol>
</section>

@endsection
