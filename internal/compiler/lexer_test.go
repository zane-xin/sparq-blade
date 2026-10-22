package compiler

import "testing"

func TestLexerTokenize(t *testing.T) {
	source := "Hello {{ name }}\n{!! html !!}\n{{-- note --}}\n@if(ok)yes@endif"

	tokens, err := NewLexer(source).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	if len(tokens) != 11 {
		t.Fatalf("unexpected token count: got %d want 11", len(tokens))
	}

	if tokens[0].Kind != TokenKindText || tokens[0].Literal != "Hello " {
		t.Fatalf("unexpected first token: %#v", tokens[0])
	}

	if tokens[1].Kind != TokenKindEchoEscaped || tokens[1].Literal != "name" {
		t.Fatalf("unexpected escaped echo token: %#v", tokens[1])
	}

	if tokens[3].Kind != TokenKindEchoRaw || tokens[3].Literal != "html" {
		t.Fatalf("unexpected raw echo token: %#v", tokens[3])
	}

	if tokens[5].Kind != TokenKindComment || tokens[5].Literal != "note" {
		t.Fatalf("unexpected comment token: %#v", tokens[5])
	}

	if tokens[7].Kind != TokenKindDirective || tokens[7].Name != "if" || tokens[7].Expression != "ok" {
		t.Fatalf("unexpected if token: %#v", tokens[7])
	}

	if tokens[7].Span.Start.Line != 4 {
		t.Fatalf("unexpected if token line: got %d want 4", tokens[7].Span.Start.Line)
	}

	if tokens[8].Kind != TokenKindText || tokens[8].Literal != "yes" {
		t.Fatalf("unexpected inline text token: %#v", tokens[8])
	}

	if tokens[10].Kind != TokenKindEOF {
		t.Fatalf("unexpected eof token: %#v", tokens[10])
	}
}

func TestLexerTokenizeNamespacedDirective(t *testing.T) {
	tokens, err := NewLexer(`@sq:snippet(name)content@sq:endsnippet`).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	if tokens[0].Kind != TokenKindDirective || tokens[0].Name != "sq:snippet" || tokens[0].Expression != "name" {
		t.Fatalf("unexpected start directive: %#v", tokens[0])
	}
	if tokens[2].Kind != TokenKindDirective || tokens[2].Name != "sq:endsnippet" {
		t.Fatalf("unexpected end directive: %#v", tokens[2])
	}
}

func TestLexerTokenizeCustomSyntax(t *testing.T) {
	source := "Hello [[ name ]]\n[[! html !]]\n[[-- note --]]\n@if(ok)yes@endif"

	tokens, err := NewLexer(source, WithSyntax(Syntax{
		EchoOpen:     "[[",
		EchoClose:    "]]",
		RawOpen:      "[[!",
		RawClose:     "!]]",
		CommentOpen:  "[[--",
		CommentClose: "--]]",
	})).Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	if len(tokens) != 11 {
		t.Fatalf("unexpected token count: got %d want 11", len(tokens))
	}

	if tokens[1].Kind != TokenKindEchoEscaped || tokens[1].Literal != "name" {
		t.Fatalf("unexpected escaped echo token: %#v", tokens[1])
	}

	if tokens[3].Kind != TokenKindEchoRaw || tokens[3].Literal != "html" {
		t.Fatalf("unexpected raw echo token: %#v", tokens[3])
	}

	if tokens[5].Kind != TokenKindComment || tokens[5].Literal != "note" {
		t.Fatalf("unexpected comment token: %#v", tokens[5])
	}
}

func TestLexerTokenizeUnclosedDirectiveExpression(t *testing.T) {
	source := "@if((value)"

	_, err := NewLexer(source).Tokenize()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
