package lexer

import (
	"fmt"
)

type TokenType uint

const (
	TOKEN_INVALID TokenType = iota
	TOKEN_EPSILON
	TOKEN_TERMINAL
	TOKEN_NOT_TERMINAL
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_OR
	TOKEN_ASTERISK
	TOKEN_PLUS
	TOKEN_POW
)

type Token interface {
	GetTokenType() TokenType
	GetSymbol() rune
	String() string
	IsCompose() bool
	GetCompose() *ComposeToken
}

type SimpleToken struct {
	Type   TokenType
	Symbol rune
	Col    int
}

func (t SimpleToken) GetTokenType() TokenType {
	return t.Type
}

func (t SimpleToken) GetSymbol() rune {
	return t.Symbol
}

func (t SimpleToken) String() string {
	switch t.Type {
	case TOKEN_EPSILON:
		return "EPSILON"
	case TOKEN_TERMINAL:
		return fmt.Sprintf("TERMINAL('%c')", t.Symbol)
	case TOKEN_NOT_TERMINAL:
		return fmt.Sprintf("NOT_TERMINAL('%c')", t.Symbol)
	case TOKEN_LPAREN:
		return "LPAREN"
	case TOKEN_RPAREN:
		return "RPAREN"
	case TOKEN_OR:
		return "OR"
	case TOKEN_ASTERISK:
		return "ASTERISK"
	case TOKEN_PLUS:
		return "PLUS"
	case TOKEN_POW:
		return "POW"
	default:
		return "INVALID"
	}
}

func (t SimpleToken) IsCompose() bool {
	return false
}

func (t SimpleToken) GetCompose() *ComposeToken {
	return nil
}

type ComposeToken struct {
	SimpleToken
	Value    int
	Literals []rune
}

func (t ComposeToken) String() string {
	if len(t.Literals) != 0 {
		return fmt.Sprintf("POW(%s)", string(t.Literals))
	}
	return fmt.Sprintf("POW(%d)", t.Value)
}

func (t ComposeToken) IsCompose() bool {
	return true
}

func (t ComposeToken) GetCompose() *ComposeToken {
	return &t
}

var _ Token = SimpleToken{}
var _ Token = ComposeToken{}
