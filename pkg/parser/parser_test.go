package parser

import (
	"reflect"
	"strings"
	"testing"

	"expr/pkg/lexer"
)

var (
	testVT = []rune{'a', 'b'}
	testVN = []rune{'S', 'A', 'B'}
)

func sym(r rune) *SymbolNode {
	return &SymbolNode{BaseNode: BaseNode{Type: AST_SYMBOL}, Symbol: r}
}

func cat(l, r ASTNode) *BinaryNode {
	return &BinaryNode{BaseNode: BaseNode{Type: AST_CONCAT_EXPR}, Left: l, Right: r}
}

func or(l, r ASTNode) *BinaryNode {
	return &BinaryNode{BaseNode: BaseNode{Type: AST_OR_EXPR}, Left: l, Right: r}
}

func zom(c ASTNode) *UnaryNode {
	return &UnaryNode{BaseNode: BaseNode{Type: AST_ZOM_EXPR}, Child: c}
}

func oom(c ASTNode) *UnaryNode {
	return &UnaryNode{BaseNode: BaseNode{Type: AST_OOM_EXPR}, Child: c}
}

func pow(b ASTNode, s ...rune) *PowerNode {
	return &PowerNode{BaseNode: BaseNode{Type: AST_POW_EXPR}, Base: b, Symbolic: s}
}

func rep(base ASTNode, n int) ASTNode {
	node := base
	for i := 1; i < n; i++ {
		node = cat(node, base)
	}
	return node
}

func parse(t *testing.T, src string) ASTNode {
	t.Helper()
	p := New(lexer.New(src, testVN, testVT))
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("parsing %q: %v", src, err)
	}
	return node
}

func TestParse(t *testing.T) {
	a, b := sym('a'), sym('b')

	tests := []struct {
		name string
		src  string
		want ASTNode
	}{
		{"single symbol", "a", a},
		{"concat", "ab", cat(a, b)},
		{"or", "a|b", or(a, b)},
		{"or is left-assoc", "a|b|a", or(or(a, b), a)},
		{"star", "a*", zom(a)},
		{"plus", "a+", oom(a)},
		{"power 2", "a²", rep(a, 2)},
		{"power 3", "a³", rep(a, 3)},
		{"power 0 is epsilon", "a⁰", sym('ε')},
		{"power 1 is identity", "a¹", a},
		{"power binds to last symbol", "ab²", cat(a, rep(b, 2))},
		{"group power", "(ab)²", rep(cat(a, b), 2)},
		{"symbolic power", "aⁿ", pow(a, 'n')},
		{"symbolic mixed power", "a²ⁿ", pow(a, '2', 'n')},
		{"group or power", "(a|b)²", rep(or(a, b), 2)},
		{"star then concat", "a*b", cat(zom(a), b)},
		{"star and plus", "a*b+", cat(zom(a), oom(b))},
		{"power of star", "a*²", rep(zom(a), 2)},
		{"power 22", "a²²", rep(a, 22)},
		{"non-terminals", "AB", cat(sym('A'), sym('B'))},
		{"full expression", "Aab²ⁿ | ba",
			or(cat(cat(sym('A'), a), pow(b, '2', 'n')), cat(b, a))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parse(t, tt.src)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsing %q:\n got: %v\nwant: %v", tt.src, got, tt.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		src     string
		wantErr string
	}{
		{"", "unexpected end of input"},
		{"*a", "unexpected token"},
		{"a)", "unexpected token"},
		{"(a", "expected ')'"},
		{"a|", "unexpected end of input"},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			p := New(lexer.New(tt.src, testVN, testVT))
			_, err := p.Parse()
			if err == nil {
				t.Fatalf("parsing %q: expected error", tt.src)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("parsing %q: error = %q, want substring %q", tt.src, err, tt.wantErr)
			}
		})
	}
}

func TestASTString(t *testing.T) {
	got := parse(t, "Aab²ⁿ | ba").String()
	want := "(((A · a) · (b^2n)) | (b · a))"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
