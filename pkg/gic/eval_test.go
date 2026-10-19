package gic

import (
	"errors"
	"testing"
)

var (
	testVT = []rune{'a', 'b', 'c'}
	testVN = []rune{'S', 'A', 'B'}
)

func compile(t *testing.T, p Productions, maxSteps int) *Program {
	t.Helper()
	prog, err := Compile(&GIC{VN: testVN, VT: testVT, P: p, S: 'S'}, maxSteps)
	if err != nil {
		t.Fatalf("Compile(%v): %v", p, err)
	}
	return prog
}

func TestEvalSymbolicPower(t *testing.T) {
	var prog = compile(t, Productions{'S': "a²ⁿbⁿ"}, 10000)

	accept := []string{"", "aab", "aaaabb"}
	reject := []string{"ab", "aaab", "aaaabbb", "aabbb", "a", "b"}

	for _, text := range accept {
		if err := prog.Eval(text); err != nil {
			t.Errorf("Eval(%q) = %v, want nil", text, err)
		}
	}

	for _, text := range reject {
		err := prog.Eval(text)
		if err == nil {
			t.Errorf("Eval(%q) = nil, want error", text)
		} else if !errors.Is(err, ErrNotAccepted) {
			t.Errorf("Eval(%q) = %v, want ErrNotAccepted", text, err)
		}
	}
}

func TestEvalRegular(t *testing.T) {
	var prog = compile(t, Productions{'S': "aS|b"}, 10000)

	for _, text := range []string{"b", "aab"} {
		if err := prog.Eval(text); err != nil {
			t.Errorf("Eval(%q) = %v, want nil", text, err)
		}
	}

	if err := prog.Eval("aba"); !errors.Is(err, ErrNotAccepted) {
		t.Errorf("Eval(aba) = %v, want ErrNotAccepted", err)
	}
}

func TestEvalStar(t *testing.T) {
	var prog = compile(t, Productions{'S': "(ab)*c"}, 10000)

	for _, text := range []string{"c", "ababc"} {
		if err := prog.Eval(text); err != nil {
			t.Errorf("Eval(%q) = %v, want nil", text, err)
		}
	}

	for _, text := range []string{"", "ab", "abac"} {
		if err := prog.Eval(text); !errors.Is(err, ErrNotAccepted) {
			t.Errorf("Eval(%q) = %v, want ErrNotAccepted", text, err)
		}
	}
}

func TestEvalPlus(t *testing.T) {
	var prog = compile(t, Productions{'S': "a+"}, 10000)

	if err := prog.Eval("aaa"); err != nil {
		t.Errorf("Eval(aaa) = %v, want nil", err)
	}

	if err := prog.Eval(""); !errors.Is(err, ErrNotAccepted) {
		t.Errorf("Eval(\"\") = %v, want ErrNotAccepted", err)
	}
}

func TestEvalRecursive(t *testing.T) {
	var prog = compile(t, Productions{'S': "aSb|ε"}, 10000)

	for _, text := range []string{"", "ab", "aabb"} {
		if err := prog.Eval(text); err != nil {
			t.Errorf("Eval(%q) = %v, want nil", text, err)
		}
	}

	if err := prog.Eval("aab"); !errors.Is(err, ErrNotAccepted) {
		t.Errorf("Eval(aab) = %v, want ErrNotAccepted", err)
	}
}

func TestEvalLeftRecursion(t *testing.T) {
	var prog = compile(t, Productions{'S': "Sa|a"}, 10000)

	if err := prog.Eval("aa"); !errors.Is(err, ErrEvalStepLimit) {
		t.Errorf("Eval(aa) = %v, want ErrEvalStepLimit", err)
	}
}

func TestEvalPowSingleAppearance(t *testing.T) {
	_, err := Compile(&GIC{VN: testVN, VT: testVT, P: Productions{'S': "a²ⁿ"}, S: 'S'}, 10000)
	if !errors.Is(err, ErrPowVarSingleAppearance) {
		t.Errorf("Compile = %v, want ErrPowVarSingleAppearance", err)
	}
}

func TestEvalPowUnverified(t *testing.T) {
	var prog = compile(t, Productions{'S': "a²ⁿ | aⁿbⁿ"}, 10000)

	if err := prog.Eval("ab"); err != nil {
		t.Errorf("Eval(ab) = %v, want nil", err)
	}

	if err := prog.Eval("aa"); !errors.Is(err, ErrPowUnverified) {
		t.Errorf("Eval(aa) = %v, want ErrPowUnverified", err)
	}
}

func TestCompileNoProductionForStart(t *testing.T) {
	_, err := Compile(&GIC{VN: testVN, VT: testVT, P: Productions{'A': "a"}, S: 'S'}, 10000)
	if !errors.Is(err, ErrNoProductionForStart) {
		t.Errorf("Compile = %v, want ErrNoProductionForStart", err)
	}
}

func TestEvalNonTerminals(t *testing.T) {
	var prog = compile(t, Productions{'S': "AB", 'A': "a", 'B': "b"}, 10000)

	if err := prog.Eval("ab"); err != nil {
		t.Errorf("Eval(ab) = %v, want nil", err)
	}

	if err := prog.Eval("ba"); !errors.Is(err, ErrNotAccepted) {
		t.Errorf("Eval(ba) = %v, want ErrNotAccepted", err)
	}
}
