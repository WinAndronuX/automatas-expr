package parser

import "fmt"

type ASTNodeType uint

const (
	AST_CONCAT_EXPR ASTNodeType = iota
	AST_OR_EXPR
	AST_ZOM_EXPR
	AST_OOM_EXPR
	AST_POW_EXPR
	AST_SYMBOL
)

type ASTNode interface {
	NodeType() ASTNodeType
	String() string
}

type BaseNode struct {
	Type ASTNodeType
}

func (n BaseNode) NodeType() ASTNodeType {
	return n.Type
}

type SymbolNode struct {
	BaseNode
	Symbol rune
}

func (n *SymbolNode) String() string {
	return string(n.Symbol)
}

type BinaryNode struct {
	BaseNode
	Left  ASTNode
	Right ASTNode
}

func (n *BinaryNode) String() string {
	switch n.Type {
	case AST_CONCAT_EXPR:
		return fmt.Sprintf("(%s · %s)", n.Left, n.Right)
	case AST_OR_EXPR:
		return fmt.Sprintf("(%s | %s)", n.Left, n.Right)
	}
	return "??"
}

type UnaryNode struct {
	BaseNode
	Child ASTNode
}

func (n *UnaryNode) String() string {
	switch n.Type {
	case AST_ZOM_EXPR:
		return fmt.Sprintf("(%s)*", n.Child)
	case AST_OOM_EXPR:
		return fmt.Sprintf("(%s)+", n.Child)
	}
	return "??"
}

type PowerNode struct {
	BaseNode
	Base     ASTNode
	Symbolic []rune
}

func (n *PowerNode) String() string {
	return fmt.Sprintf("(%s^%s)", n.Base, string(n.Symbolic))
}

var (
	_ ASTNode = (*SymbolNode)(nil)
	_ ASTNode = (*BinaryNode)(nil)
	_ ASTNode = (*UnaryNode)(nil)
	_ ASTNode = (*PowerNode)(nil)
)
