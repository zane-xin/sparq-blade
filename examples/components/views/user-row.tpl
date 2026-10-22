@component('user-card')
  @slot('title')
    @sq:badge(user.name)
  @endslot

  @if(user.active)
    <span class="state state-on">active</span>
  @else
    <span class="state state-off">inactive</span>
  @endif
@endcomponent
