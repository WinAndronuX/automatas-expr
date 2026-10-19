package gic

import (
	"expr/pkg/lexer"
	"expr/pkg/parser"
	"fmt"
	"slices"
)

type Program struct {
	gic      *GIC
	rules    map[rune]parser.ASTNode
	maxSteps int
}

func Compile(gic *GIC, maxSteps int) (*Program, error) {

	var prog = new(Program)

	prog.gic = gic
	prog.maxSteps = maxSteps

	if !slices.Contains(gic.VN, gic.S) {
		return nil, fmt.Errorf("%v: %c", ErrSymbolNotDeclared, gic.S)
	}

	var rules = map[rune]parser.ASTNode{}

	for nt, ru := range gic.P {

		if !slices.Contains(gic.VN, nt) {
			return nil, fmt.Errorf("%v: %c", ErrSymbolNotDeclared, nt)
		}

		var p = parser.New(lexer.New(ru, gic.VN, gic.VT))

		ast, err := p.Parse()
		if err != nil {
			return nil, err
		}

		rules[nt] = ast
	}

	if _, ok := rules[gic.S]; !ok {
		return nil, fmt.Errorf("%w: %c", ErrNoProductionForStart, gic.S)
	}

	var powVars = map[rune]int{}
	for _, ast := range rules {
		countPowVars(ast, powVars)
	}

	for v, c := range powVars {
		if c < 2 {
			return nil, fmt.Errorf("%w: %c appears in %d pow expression(s)", ErrPowVarSingleAppearance, v, c)
		}
	}

	prog.rules = rules

	return prog, nil
}

func countPowVars(node parser.ASTNode, counts map[rune]int) {
	switch n := node.(type) {
	case *parser.BinaryNode:
		countPowVars(n.Left, counts)
		countPowVars(n.Right, counts)
	case *parser.UnaryNode:
		countPowVars(n.Child, counts)
	case *parser.PowerNode:
		for _, r := range n.Symbolic {
			if r >= 'a' && r <= 'z' {
				counts[r]++
			}
		}
	}
}
