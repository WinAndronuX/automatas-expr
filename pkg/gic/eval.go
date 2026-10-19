package gic

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"expr/pkg/parser"
)

type pdaSym struct {
	isMarker bool
	node     parser.ASTNode
	child    parser.ASTNode
}

type pdaEval struct {
	prog       *Program
	input      []rune
	steps      int
	visited    map[string]bool
	verified   bool
	unverified bool
}

func (p *Program) Eval(text string) error {

	rule, ok := p.rules[p.gic.S]
	if !ok {
		return fmt.Errorf("%v: %c", ErrNoProductionForStart, p.gic.S)
	}

	var e = &pdaEval{
		prog:    p,
		input:   []rune(text),
		visited: map[string]bool{},
	}

	e.run(0, []pdaSym{{node: rule}}, map[rune]int{}, map[rune]bool{})

	if e.verified {
		return nil
	}

	if e.steps > p.maxSteps {
		return fmt.Errorf("%w (max %d)", ErrEvalStepLimit, p.maxSteps)
	}

	if e.unverified {
		return ErrPowUnverified
	}

	return fmt.Errorf("%w: %q", ErrNotAccepted, text)
}

func (e *pdaEval) run(pos int, stack []pdaSym, subst map[rune]int, checked map[rune]bool) {

	if e.verified || e.steps > e.prog.maxSteps {
		return
	}

	e.steps++

	var key = configKey(pos, stack, subst, checked)

	if e.visited[key] {
		return
	}
	e.visited[key] = true

	if len(stack) == 0 {
		if pos == len(e.input) {

			v := allChecked(subst, checked)

			e.verified = v
			e.unverified = !v
		}
		return
	}

	var top = stack[len(stack)-1]
	var rest = stack[:len(stack)-1]

	if top.isMarker {

		e.run(pos, rest, subst, checked)
		e.run(pos, pushMarker(rest, top.child), subst, checked)

		return
	}

	switch n := top.node.(type) {
	case *parser.SymbolNode:
		e.runSymbol(n, pos, rest, subst, checked)
	case *parser.BinaryNode:
		e.runBinary(n, pos, rest, subst, checked)
	case *parser.UnaryNode:
		e.runUnary(n, pos, rest, subst, checked)
	case *parser.PowerNode:
		e.runPower(n, pos, rest, subst, checked)
	}
}

func (e *pdaEval) runSymbol(n *parser.SymbolNode, pos int, rest []pdaSym, subst map[rune]int, checked map[rune]bool) {

	switch {
	case n.Symbol == 'ε':
		e.run(pos, rest, subst, checked)
	case slices.Contains(e.prog.gic.VN, n.Symbol):
		e.run(pos, pushNode(rest, e.prog.rules[n.Symbol]), subst, checked)
	case pos < len(e.input) && e.input[pos] == n.Symbol:
		e.run(pos+1, rest, subst, checked)
	}
}

func (e *pdaEval) runBinary(n *parser.BinaryNode, pos int, rest []pdaSym, subst map[rune]int, checked map[rune]bool) {

	switch n.NodeType() {
	case parser.AST_CONCAT_EXPR:
		e.run(pos, pushNode(pushNode(rest, n.Right), n.Left), subst, checked)
	case parser.AST_OR_EXPR:
		e.run(pos, pushNode(rest, n.Left), subst, checked)
		e.run(pos, pushNode(rest, n.Right), subst, checked)
	}
}

func (e *pdaEval) runUnary(n *parser.UnaryNode, pos int, rest []pdaSym, subst map[rune]int, checked map[rune]bool) {

	switch n.NodeType() {
	case parser.AST_ZOM_EXPR:
		e.run(pos, rest, subst, checked)
		e.run(pos, pushMarker(rest, n.Child), subst, checked)
	case parser.AST_OOM_EXPR:
		e.run(pos, pushMarker(rest, n.Child), subst, checked)
	}
}

func (e *pdaEval) runPower(n *parser.PowerNode, pos int, rest []pdaSym, subst map[rune]int, checked map[rune]bool) {

	var factors = parseFactors(n.Symbolic)

	var prod = 1
	var unbound []rune

	for _, f := range factors {
		if f.isVar {
			if v, ok := subst[f.name]; ok {
				prod *= v
			} else {
				unbound = append(unbound, f.name)
			}
		} else {
			prod *= f.val
		}
	}

	if len(unbound) > 1 {
		return
	}

	if len(unbound) == 0 {
		var newChecked = map[rune]bool{}

		maps.Copy(newChecked, checked)

		for _, f := range factors {

			if f.isVar {
				newChecked[f.name] = true
			}
		}

		e.pushCopies(n.Base, prod, pos, rest, subst, newChecked)
		return
	}

	if prod == 0 {
		return
	}

	var v = unbound[0]
	var remaining = len(e.input) - pos
	for k := 0; prod*k <= remaining; k++ {
		var newSubst = map[rune]int{}

		maps.Copy(newSubst, subst)

		newSubst[v] = k
		e.pushCopies(n.Base, prod*k, pos, rest, newSubst, checked)
	}
}

func (e *pdaEval) pushCopies(base parser.ASTNode, count, pos int, rest []pdaSym, subst map[rune]int, checked map[rune]bool) {
	var stack = rest

	for i := 0; i < count; i++ {
		stack = pushNode(stack, base)
	}

	e.run(pos, stack, subst, checked)
}

func pushNode(stack []pdaSym, node parser.ASTNode) []pdaSym {
	return append(append([]pdaSym{}, stack...), pdaSym{node: node})
}

func pushMarker(stack []pdaSym, child parser.ASTNode) []pdaSym {
	stack = append(stack, pdaSym{isMarker: true, child: child})
	return append(stack, pdaSym{node: child})
}

type powFactor struct {
	val   int
	isVar bool
	name  rune
}

func parseFactors(run []rune) []powFactor {
	var factors []powFactor
	for i := 0; i < len(run); {
		if run[i] >= '0' && run[i] <= '9' {
			var val = 0
			for i < len(run) && run[i] >= '0' && run[i] <= '9' {
				val = val*10 + int(run[i]-'0')
				i++
			}
			factors = append(factors, powFactor{val: val})
		} else {
			factors = append(factors, powFactor{isVar: true, name: run[i]})
			i++
		}
	}
	return factors
}

func allChecked(subst map[rune]int, checked map[rune]bool) bool {
	for k := range subst {
		if !checked[k] {
			return false
		}
	}
	return true
}

func configKey(pos int, stack []pdaSym, subst map[rune]int, checked map[rune]bool) string {

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d|", pos)

	for _, f := range stack {
		if f.isMarker {
			fmt.Fprintf(&sb, "M%p;", f.child)
		} else {
			fmt.Fprintf(&sb, "%p;", f.node)
		}
	}

	sb.WriteByte('|')

	var keys = make([]rune, 0, len(subst))
	for k := range subst {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, k := range keys {
		fmt.Fprintf(&sb, "%c=%d;", k, subst[k])
	}

	sb.WriteByte('|')

	var ckeys = make([]rune, 0, len(checked))

	for k := range checked {
		ckeys = append(ckeys, k)
	}

	slices.Sort(ckeys)

	for _, k := range ckeys {
		fmt.Fprintf(&sb, "%c;", k)
	}

	return sb.String()
}
