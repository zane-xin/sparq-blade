package blade_test

import (
	"testing"

	blade "github.com/zane-xin/sparq-blade"
)

func FuzzLex(f *testing.F) {
	for _, seed := range fuzzSeeds() {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, source string) {
		_, _ = blade.Lex(source)
	})
}

func FuzzParse(f *testing.F) {
	for _, seed := range fuzzSeeds() {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, source string) {
		_, _ = blade.Parse(source)
	})
}

func fuzzSeeds() []string {
	return []string{
		"",
		"plain text",
		"<h1>{{ title }}</h1>",
		"{!! html !!}",
		"{{-- comment --}}",
		"@if(ok)yes@else no@endif",
		"@foreach(items as item){{ item.name }}@endforeach",
		"@for(i in 1..3){{ i }}@endfor",
		"@list(items, as=item){{ item.name }}@empty none@endlist",
		"@list(items, as=item)@children(as=child, wrap=\"ul\"){{ child.name }}@endchildren@endlist",
		"@extends('layout')@section('content')body@endsection",
		"@include('partial', data)",
		"@component('card')body@slot('title')Title@endslot@endcomponent",
		"@sq:arc(limit=1, as=article)<span>{{ article.title }}</span>@end",
		"@if(",
		"{{ $",
		"@component('card')",
	}
}
