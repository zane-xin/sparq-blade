<nav>
  @sq:nav(as=item)
    @if(item.active)
      <a class="active" href="{{ item.url }}">{{ item.title }}</a>
    @else
      <a href="{{ item.url }}">{{ item.title }}</a>
    @endif
  @end
</nav>
