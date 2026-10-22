@extends('layout')

@section('content')
<h1>{{ title }}</h1>
<p>Current path: {{ path }}</p>
<p>Server time: {{ now }}</p>

@sq:panel(features)
  <ul>
    @list(features, as=feature)
    <li>
      @if(feature.ready)
        {{ feature.name }}: ready
      @else
        {{ feature.name }}: pending
      @endif
    </li>
    @end
  </ul>
@end
@endsection
