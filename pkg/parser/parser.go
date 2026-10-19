package parser

import (
	"errors"
	"fmt"
	"io"

	"expr/pkg/lexer"
)

const (
	bpOr      = 1
	bpConcat  = 2
	bpPostfix = 3
)

type Parser struct {
	lx  *lexer.Lexer
	cur lexer.Token
	err error
}

func New(lx *lexer.Lexer) *Parser {
	return &Parser{lx: lx}
}

func (p *Parser) Parse() (ASTNode, error) {
	p.next()

	node, err := p.parseExpr(0)
	if err != nil {
		return nil, err
	}

	if p.err != nil {
		return nil, p.err
	}

	if p.cur != nil {
		return nil, fmt.Errorf("unexpected token %v after expression", p.cur)
	}

	return node, nil
}

func (p *Parser) next() {
	if p.err != nil {
		return
	}

	tok, err := p.lx.NextToken()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			p.err = err
		}
		p.cur = nil
		return
	}

	p.cur = tok
}

func (p *Parser) parseExpr(minPrec int) (ASTNode, error) {

	left, err := p.parsePrefix()
	if err != nil {
		return nil, err
	}

	for p.cur != nil {
		var bp int

		switch p.cur.GetTokenType() {
		case lexer.TOKEN_OR:
			bp = bpOr
		case lexer.TOKEN_ASTERISK, lexer.TOKEN_PLUS, lexer.TOKEN_POW:
			bp = bpPostfix
		case lexer.TOKEN_TERMINAL, lexer.TOKEN_NOT_TERMINAL, lexer.TOKEN_EPSILON, lexer.TOKEN_LPAREN:
			bp = bpConcat
		default:
			bp = 0
		}

		if bp == 0 || bp < minPrec {
			break
		}

		switch p.cur.GetTokenType() {
		case lexer.TOKEN_OR:
			p.next()
			right, err := p.parseExpr(bpOr + 1)
			if err != nil {
				return nil, err
			}
			left = &BinaryNode{BaseNode: BaseNode{Type: AST_OR_EXPR}, Left: left, Right: right}

		case lexer.TOKEN_ASTERISK:
			p.next()
			left = &UnaryNode{BaseNode: BaseNode{Type: AST_ZOM_EXPR}, Child: left}

		case lexer.TOKEN_PLUS:
			p.next()
			left = &UnaryNode{BaseNode: BaseNode{Type: AST_OOM_EXPR}, Child: left}

		case lexer.TOKEN_POW:
			ct := p.cur.GetCompose()
			if ct == nil {
				return nil, fmt.Errorf("invalid power token")
			}
			p.next()
			left = applyPower(left, ct)

		default:
			right, err := p.parseExpr(bpConcat + 1)
			if err != nil {
				return nil, err
			}
			left = &BinaryNode{BaseNode: BaseNode{Type: AST_CONCAT_EXPR}, Left: left, Right: right}
		}
	}

	return left, nil
}

func (p *Parser) parsePrefix() (ASTNode, error) {

	tok := p.cur
	if tok == nil {
		return nil, fmt.Errorf("unexpected end of input")
	}

	switch tok.GetTokenType() {
	case lexer.TOKEN_TERMINAL, lexer.TOKEN_NOT_TERMINAL, lexer.TOKEN_EPSILON:
		p.next()
		return &SymbolNode{BaseNode: BaseNode{Type: AST_SYMBOL}, Symbol: tok.GetSymbol()}, nil

	case lexer.TOKEN_LPAREN:
		p.next()
		node, err := p.parseExpr(0)
		if err != nil {
			return nil, err
		}
		if p.cur == nil || p.cur.GetTokenType() != lexer.TOKEN_RPAREN {
			return nil, fmt.Errorf("expected ')'")
		}
		p.next()
		return node, nil

	default:
		return nil, fmt.Errorf("unexpected token %v", tok)
	}
}

func applyPower(base ASTNode, ct *lexer.ComposeToken) ASTNode {

	if len(ct.Literals) != 0 {
		return &PowerNode{BaseNode: BaseNode{Type: AST_POW_EXPR}, Base: base, Symbolic: ct.Literals}
	}

	switch ct.Value {
	case 0:
		return &SymbolNode{BaseNode: BaseNode{Type: AST_SYMBOL}, Symbol: 'ε'}
	case 1:
		return base
	default:
		node := base
		for i := 1; i < ct.Value; i++ {
			node = &BinaryNode{BaseNode: BaseNode{Type: AST_CONCAT_EXPR}, Left: node, Right: base}
		}
		return node
	}
}
