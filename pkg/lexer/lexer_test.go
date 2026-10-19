package lexer

import (
	"errors"
	"io"
	"reflect"
	"testing"
)

var (
	testVT = []rune{'a', 'b'}
	testVN = []rune{'A'}
)

func lexAll(t *testing.T, src string) []Token {
	t.Helper()
	lx := New(src, testVN, testVT)
	var toks []Token
	for {
		tok, err := lx.NextToken()
		if errors.Is(err, io.EOF) {
			return toks
		}
		if err != nil {
			t.Fatalf("lexing %q: unexpected error: %v", src, err)
		}
		if tok == nil {
			t.Fatalf("lexing %q: nil token without error", src)
		}
		toks = append(toks, tok)
		if len(toks) > 100 {
			t.Fatalf("lexing %q: too many tokens (runaway loop)", src)
		}
	}
}

func TestLexTokens(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []Token
	}{
		{
			"single char",
			"a",
			[]Token{&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0}},
		},
		{
			"concat",
			"ab",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'b', Col: 1},
			},
		},
		{
			"star and plus",
			"a*b+",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0},
				&SimpleToken{Type: TOKEN_ASTERISK, Symbol: '*', Col: 1},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'b', Col: 2},
				&SimpleToken{Type: TOKEN_PLUS, Symbol: '+', Col: 3},
			},
		},
		{
			"leading spaces",
			"  ab",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 2},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'b', Col: 3},
			},
		},
		{
			"superscript digit",
			"a²",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0},
				&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: '²', Col: 1}, Value: 2},
			},
		},
		{
			"superscript multidigit",
			"a¹²",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0},
				&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: '¹', Col: 1}, Value: 12},
			},
		},
		{
			"superscript then char",
			"a²b",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0},
				&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: '²', Col: 1}, Value: 2},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'b', Col: 2},
			},
		},
		{
			"group power",
			"(ab)²",
			[]Token{
				&SimpleToken{Type: TOKEN_LPAREN, Symbol: '(', Col: 0},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 1},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'b', Col: 2},
				&SimpleToken{Type: TOKEN_RPAREN, Symbol: ')', Col: 3},
				&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: '²', Col: 4}, Value: 2},
			},
		},
		{
			"superscript symbolic letter",
			"aⁿ",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0},
				&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: 'ⁿ', Col: 1}, Literals: []rune{'n'}},
			},
		},
		{
			"power binds to last symbol",
			"ab²",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'b', Col: 1},
				&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: '²', Col: 2}, Value: 2},
			},
		},
		{
			"or operator",
			"a|b",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0},
				&SimpleToken{Type: TOKEN_OR, Symbol: '|', Col: 1},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'b', Col: 2},
			},
		},
		{
			"superscript mixed symbolic",
			"a²ⁿ",
			[]Token{
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 0},
				&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: '²', Col: 1}, Literals: []rune{'2', 'n'}},
			},
		},
		{
			"full expression",
			"Aab²ⁿ | ba",
			[]Token{
				&SimpleToken{Type: TOKEN_NOT_TERMINAL, Symbol: 'A', Col: 0},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 1},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'b', Col: 2},
				&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: '²', Col: 3}, Literals: []rune{'2', 'n'}},
				&SimpleToken{Type: TOKEN_OR, Symbol: '|', Col: 6},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'b', Col: 8},
				&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a', Col: 9},
			},
		},
		{
			"epsilon",
			"ε",
			[]Token{&SimpleToken{Type: TOKEN_EPSILON, Symbol: 'ε', Col: 0}},
		},
		{
			"empty input",
			"",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lexAll(t, tt.src)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lexing %q:\n got: %v\nwant: %v", tt.src, got, tt.want)
			}
		})
	}
}

func TestLexInvalid(t *testing.T) {
	lx := New("a!", testVN, testVT)
	var err error
	for i := 0; i < 10; i++ {
		_, err = lx.NextToken()
		if err != nil {
			break
		}
	}
	if err == nil {
		t.Fatal("expected error for invalid char '!'")
	}
}

func TestComposeTokenString(t *testing.T) {
	num := (&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: '²'}, Value: 2}).String()
	if num != "POW(2)" {
		t.Errorf("ComposeToken numeric String() = %q, want %q", num, "POW(2)")
	}
	lit := (&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: 'ⁿ'}, Literals: []rune{'n'}}).String()
	if lit != "POW(n)" {
		t.Errorf("ComposeToken symbolic String() = %q, want %q", lit, "POW(n)")
	}
	mix := (&ComposeToken{SimpleToken: SimpleToken{Type: TOKEN_POW, Symbol: '²'}, Literals: []rune{'2', 'n'}}).String()
	if mix != "POW(2n)" {
		t.Errorf("ComposeToken mixed String() = %q, want %q", mix, "POW(2n)")
	}
	if (&SimpleToken{Type: TOKEN_TERMINAL, Symbol: 'a'}).String() != "TERMINAL('a')" {
		t.Error("SimpleToken String() broke")
	}
}
