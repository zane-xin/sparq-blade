@extends('layout')

@section('content')
<h1>{{ title }}</h1>
<ul>
  @list(items, as=item)
  <li>{{ item.name }}</li>
  @end
</ul>
@endsection
