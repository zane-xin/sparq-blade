<nav class="crumb">
  @list(crumb, as=item)
    <a href="{{ item.URL }}">{{ item.Title }}</a>
  @empty
    <span>首页</span>
  @end
</nav>
