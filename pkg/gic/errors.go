package gic

import "errors"

var (
	ErrSymbolNotDeclared      = errors.New("Symbol not declared")
	ErrNoProductionForStart   = errors.New("No production for start symbol")
	ErrNotAccepted            = errors.New("Text not accepted")
	ErrEvalStepLimit          = errors.New("Evaluation step limit exceeded (left recursion?)")
	ErrPowUnverified          = errors.New("Symbolic power variable never verified")
	ErrPowVarSingleAppearance = errors.New("Symbolic power variable appears only once")
)
