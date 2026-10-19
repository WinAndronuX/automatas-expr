package lexer

import (
	"fmt"
	"io"
	"slices"
	"strconv"
)

type Lexer struct {
	Source []rune
	vT     []rune
	vN     []rune
	cursor int
	peek   int
	ch     rune
	col    int
}

func New(src string, vN, vT []rune) *Lexer {
	return &Lexer{Source: []rune(src), cursor: 0, peek: 0, vN: vN, vT: vT, col: 0}
}

func (lx *Lexer) NextToken() (Token, error) {

	if !lx.readChar() {
		return nil, io.EOF
	}

	lx.skipWhiteSpaces()

	if lx.ch == 0 {
		return nil, io.EOF
	}

	var tokenType TokenType

	switch {
	case slices.Contains(lx.vT, lx.ch):
		tokenType = TOKEN_TERMINAL
	case slices.Contains(lx.vN, lx.ch):
		tokenType = TOKEN_NOT_TERMINAL
	case lx.ch == 'ε':
		tokenType = TOKEN_EPSILON
	case lx.ch == '(':
		tokenType = TOKEN_LPAREN
	case lx.ch == ')':
		tokenType = TOKEN_RPAREN
	case lx.ch == '|':
		tokenType = TOKEN_OR
	case lx.ch == '*':
		tokenType = TOKEN_ASTERISK
	case lx.ch == '+':
		tokenType = TOKEN_PLUS
	case isSuper(lx.ch):
		return lx.consumeSuper()
	default:
		return nil, fmt.Errorf("%v: %c", ErrTokenInvalid, lx.ch)
	}

	var token = &SimpleToken{
		Type:   tokenType,
		Symbol: lx.ch,
		Col:    lx.cursor,
	}

	return token, nil
}

func (lx *Lexer) readChar() bool {

	if lx.peek >= len(lx.Source) {
		lx.ch = 0
		return false
	}

	lx.ch = lx.Source[lx.peek]

	lx.cursor = lx.peek
	lx.peek++
	lx.col++

	return true
}

func (lx *Lexer) skipWhiteSpaces() {
	for lx.ch == ' ' && lx.readChar() {
	}
}

var nsupers = map[rune]rune{
	'⁰': '0', '¹': '1', '²': '2', '³': '3', '⁴': '4',
	'⁵': '5', '⁶': '6', '⁷': '7', '⁸': '8', '⁹': '9',
}

var lsupers = map[rune]rune{
	'ᵃ': 'a', 'ᵇ': 'b', 'ᶜ': 'c', 'ᵈ': 'd', 'ᵉ': 'e',
	'ᶠ': 'f', 'ᵍ': 'g', 'ʰ': 'h', 'ⁱ': 'i', 'ʲ': 'j',
	'ᵏ': 'k', 'ˡ': 'l', 'ᵐ': 'm', 'ⁿ': 'n', 'ᵒ': 'o',
	'ᵖ': 'p', 'ʳ': 'r', 'ˢ': 's', 'ᵗ': 't', 'ᵘ': 'u',
	'ᵛ': 'v', 'ʷ': 'w', 'ˣ': 'x', 'ʸ': 'y', 'ᶻ': 'z',
}

func isNumSuper(r rune) bool {
	_, ok := nsupers[r]
	return ok
}

func isLetSuper(r rune) bool {
	_, ok := lsupers[r]
	return ok
}

func isSuper(r rune) bool {
	return isNumSuper(r) || isLetSuper(r)
}

func (lx *Lexer) consumeSuper() (Token, error) {

	var start = lx.ch
	var col = lx.cursor

	var run []rune

	for isSuper(lx.ch) {

		if isLetSuper(lx.ch) {
			run = append(run, lsupers[lx.ch])
		} else {
			run = append(run, nsupers[lx.ch])
		}

		if !lx.readChar() {
			break
		}
	}

	if lx.ch != 0 {
		lx.peek = lx.cursor
	}

	var token = &ComposeToken{
		SimpleToken: SimpleToken{
			Type:   TOKEN_POW,
			Symbol: start,
			Col:    col,
		},
	}

	var allDigits = true
	for _, r := range run {
		if r < '0' || r > '9' {
			allDigits = false
			break
		}
	}

	if allDigits {
		var num, errConv = strconv.Atoi(string(run))
		if errConv != nil {
			return nil, errConv
		}

		token.Value = num
	} else {
		token.Literals = run
	}

	return token, nil
}
