package compiler

import (
	"strings"
	"testing"
)

func TestParserParseConditional(t *testing.T) {
	source := "@if(ok)A@elseif(other)B@else C@endif"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	document, err := NewParser(tokens).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(document.Nodes) != 1 {
		t.Fatalf("unexpected root node count: got %d want 1", len(document.Nodes))
	}

	node, ok := document.Nodes[0].(*ConditionalNode)
	if !ok {
		t.Fatalf("unexpected node type: %T", document.Nodes[0])
	}

	if len(node.Branches) != 2 {
		t.Fatalf("unexpected branch count: got %d want 2", len(node.Branches))
	}

	if node.Branches[0].Directive != "if" || node.Branches[0].Expression != "ok" {
		t.Fatalf("unexpected first branch: %#v", node.Branches[0])
	}

	if len(node.ElseBody) != 1 {
		t.Fatalf("unexpected else body count: got %d want 1", len(node.ElseBody))
	}
}

func TestParserParseLoopAndInclude(t *testing.T) {
	source := "@foreach(items as item){{ item }}@endforeach@include('footer')"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	document, err := NewParser(tokens).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(document.Nodes) != 2 {
		t.Fatalf("unexpected root node count: got %d want 2", len(document.Nodes))
	}

	loop, ok := document.Nodes[0].(*LoopNode)
	if !ok {
		t.Fatalf("unexpected first node type: %T", document.Nodes[0])
	}

	if loop.Directive != "foreach" || loop.Expression != "items as item" {
		t.Fatalf("unexpected loop node: %#v", loop)
	}

	if len(loop.Body) != 1 {
		t.Fatalf("unexpected loop body count: got %d want 1", len(loop.Body))
	}

	include, ok := document.Nodes[1].(*IncludeNode)
	if !ok {
		t.Fatalf("unexpected second node type: %T", document.Nodes[1])
	}

	if include.Target != "'footer'" {
		t.Fatalf("unexpected include target: %#v", include.Target)
	}
}

func TestParserParseChildren(t *testing.T) {
	source := `@children(as=child, wrap="ul"){{ child.title }}@empty none@endchildren`

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	document, err := NewParser(tokens).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(document.Nodes) != 1 {
		t.Fatalf("unexpected root node count: got %d want 1", len(document.Nodes))
	}

	node, ok := document.Nodes[0].(*ChildrenNode)
	if !ok {
		t.Fatalf("unexpected node type: %T", document.Nodes[0])
	}

	if node.Expression != `as=child, wrap="ul"` {
		t.Fatalf("unexpected children expression: %#v", node.Expression)
	}
	if len(node.Body) != 1 || len(node.EmptyBody) != 1 {
		t.Fatalf("unexpected children bodies: %#v", node)
	}
}

func TestParserParseGenericEndDirectives(t *testing.T) {
	source := `@if(ok)A@else B@end @list(items, as=item){{ item.title }}@empty none@end @children(as=child){{ child.title }}@end`

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	document, err := NewParser(tokens).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var conditionals, lists, children int
	for _, node := range document.Nodes {
		switch node.(type) {
		case *ConditionalNode:
			conditionals++
		case *ListNode:
			lists++
		case *ChildrenNode:
			children++
		}
	}
	if conditionals != 1 || lists != 1 || children != 1 {
		t.Fatalf("unexpected nodes: conditionals=%d lists=%d children=%d", conditionals, lists, children)
	}
}

func TestParserParseSection(t *testing.T) {
	source := "@section('content')Hello @yield('body')@endsection"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	document, err := NewParser(tokens).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(document.Nodes) != 1 {
		t.Fatalf("unexpected root node count: got %d want 1", len(document.Nodes))
	}

	section, ok := document.Nodes[0].(*SectionNode)
	if !ok {
		t.Fatalf("unexpected node type: %T", document.Nodes[0])
	}

	if section.Name != "'content'" {
		t.Fatalf("unexpected section name: %#v", section.Name)
	}

	if len(section.Body) != 2 {
		t.Fatalf("unexpected section body count: got %d want 2", len(section.Body))
	}
}

func TestParserParseComponentUnclosed(t *testing.T) {
	source := "@component('card')Hello"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	_, err = NewParser(tokens).Parse()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "component 未正常闭合") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParserParseSlotUnclosed(t *testing.T) {
	source := "@component('card')@slot('title')Hello@endcomponent"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	_, err = NewParser(tokens).Parse()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "slot 未正常闭合") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParserParseCustomDirectiveBlockWithGenericEnd(t *testing.T) {
	source := "@demo(limit=1, as=item)Body@end"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	document, err := NewParser(tokens, WithCustomDirectiveSpecs(CustomDirectiveSpec{
		Name:        "demo",
		AllowSingle: false,
		EndName:     "enddemo",
	})).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(document.Nodes) != 1 {
		t.Fatalf("unexpected root node count: got %d want 1", len(document.Nodes))
	}
	block, ok := document.Nodes[0].(*CustomDirectiveNode)
	if !ok {
		t.Fatalf("unexpected node type: %T", document.Nodes[0])
	}
	if block.Name != "demo" || block.EndName != "enddemo" || len(block.Body) != 1 {
		t.Fatalf("unexpected block directive node: %#v", block)
	}
}

func TestParserParseCustomDirectiveEmptyBranch(t *testing.T) {
	source := "@sq:demo(as=item)Body@empty Empty@end"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	document, err := NewParser(tokens, WithCustomDirectiveSpecs(CustomDirectiveSpec{
		Name:        "sq:demo",
		AllowSingle: false,
		AllowEmpty:  true,
		EndName:     "sq:enddemo",
	})).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	block, ok := document.Nodes[0].(*CustomDirectiveNode)
	if !ok {
		t.Fatalf("unexpected node type: %T", document.Nodes[0])
	}
	if len(block.Body) != 1 || len(block.EmptyBody) != 1 {
		t.Fatalf("unexpected block branches: %#v", block)
	}
}
func TestParserParseCustomDirectiveOutStaysSingleBeforeGenericEnd(t *testing.T) {
	source := "@demo(out=page)@list(page.items, as=item){{ item.title }}@end"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	document, err := NewParser(tokens, WithCustomDirectiveSpecs(CustomDirectiveSpec{
		Name:        "demo",
		AllowSingle: true,
		EndName:     "enddemo",
	})).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(document.Nodes) != 2 {
		t.Fatalf("unexpected root node count: got %d want 2", len(document.Nodes))
	}
	single, ok := document.Nodes[0].(*CustomDirectiveNode)
	if !ok {
		t.Fatalf("unexpected first node type: %T", document.Nodes[0])
	}
	if single.Name != "demo" || single.EndName != "" {
		t.Fatalf("unexpected single directive node: %#v", single)
	}
	if _, ok := document.Nodes[1].(*ListNode); !ok {
		t.Fatalf("unexpected second node type: %T", document.Nodes[1])
	}
}

func TestParserParseCustomDirectiveSingleAndBlock(t *testing.T) {
	source := "@demo(limit=1)@demo(limit=2)Body@enddemo"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	document, err := NewParser(tokens, WithCustomDirectiveSpecs(CustomDirectiveSpec{
		Name:        "demo",
		AllowSingle: true,
		EndName:     "enddemo",
	})).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(document.Nodes) != 2 {
		t.Fatalf("unexpected root node count: got %d want 2", len(document.Nodes))
	}

	single, ok := document.Nodes[0].(*CustomDirectiveNode)
	if !ok {
		t.Fatalf("unexpected first node type: %T", document.Nodes[0])
	}
	if single.Name != "demo" || single.EndName != "" {
		t.Fatalf("unexpected single directive node: %#v", single)
	}

	block, ok := document.Nodes[1].(*CustomDirectiveNode)
	if !ok {
		t.Fatalf("unexpected second node type: %T", document.Nodes[1])
	}
	if block.Name != "demo" || block.EndName != "enddemo" {
		t.Fatalf("unexpected block directive node: %#v", block)
	}
	if len(block.Body) != 1 {
		t.Fatalf("unexpected block body count: got %d want 1", len(block.Body))
	}
}
