@extends('layout')

@section('content')
<h1>{{ title }}</h1>
<p>Current path: {{ path }}</p>

@sq:panel(users)
  <div class="irid">
    @each('user-row', users, 'user', 'empty-users')
  </div>
@end
@endsection
