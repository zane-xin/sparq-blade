package runtime

import (
	"errors"
	"testing"
)

func TestResolveExpressionShortCircuitsLogicalOperators(t *testing.T) {
	calls := 0
	scope := NewScope(nil).WithFunctions(map[string]Function{
		"plugin.fail": func(...any) (any, error) {
			calls++
			return nil, errors.New("should not run")
		},
	})

	tests := []struct {
		expression string
		want       bool
	}{
		{expression: "true || plugin.fail()", want: true},
		{expression: "false && plugin.fail()", want: false},
	}

	for _, test := range tests {
		value, err := ResolveExpression(scope, test.expression)
		if err != nil {
			t.Fatalf("ResolveExpression(%q) error = %v", test.expression, err)
		}
		if value != test.want {
			t.Fatalf("ResolveExpression(%q) = %#v, want %t", test.expression, value, test.want)
		}
	}
	if calls != 0 {
		t.Fatalf("short-circuited function calls = %d, want 0", calls)
	}
}

func TestResolveExpressionValidatesSkippedBranchSyntax(t *testing.T) {
	_, err := ResolveExpression(NewScope(nil), "true || (false")
	if err == nil {
		t.Fatal("expected syntax error, got nil")
	}
}

func TestResolveExpressionComparesLargeIntegersExactly(t *testing.T) {
	scope := NewScope(map[string]any{
		"left":     uint64(9007199254740992),
		"right":    uint64(9007199254740993),
		"negative": int64(-1),
	})

	tests := []struct {
		expression string
		want       bool
	}{
		{expression: "left == right", want: false},
		{expression: "left != right", want: true},
		{expression: "left < right", want: true},
		{expression: "right > left", want: true},
		{expression: "negative < left", want: true},
	}

	for _, test := range tests {
		value, err := ResolveExpression(scope, test.expression)
		if err != nil {
			t.Fatalf("ResolveExpression(%q) error = %v", test.expression, err)
		}
		if value != test.want {
			t.Fatalf("ResolveExpression(%q) = %#v, want %t", test.expression, value, test.want)
		}
	}
}
