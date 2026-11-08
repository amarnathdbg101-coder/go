package expr

import (
	"sync"
	"testing"
)

func TestExpr_Arithmetic(t *testing.T) {
	tests := []struct {
		expr     string
		env      map[string]any
		expected float64
	}{
		{"10 + 20 * 3", nil, 70},
		{"(10 + 20) * 3", nil, 90},
		{"100 / 4 - 5", nil, 20},
		{"17 % 5", nil, 2},
		{"-5 + 12", nil, 7},
		{"price * quantity * (1 - discount)", map[string]any{"price": 100, "quantity": 5, "discount": 0.2}, 400},
	}

	for _, tc := range tests {
		val, err := EvalFloat(tc.expr, tc.env)
		if err != nil {
			t.Fatalf("expression %s failed: %v", tc.expr, err)
		}
		if val != tc.expected {
			t.Fatalf("expression %s: expected %f, got %f", tc.expr, tc.expected, val)
		}
	}
}

func TestExpr_BooleanAndLogic(t *testing.T) {
	env := map[string]any{
		"age":    24,
		"active": true,
		"role":   "admin",
		"score":  85,
	}

	tests := []struct {
		expr     string
		expected bool
	}{
		{"age >= 18 && active", true},
		{"role == 'admin' || score > 90", true},
		{"role != 'admin' || score < 50", false},
		{"!active", false},
		{"(score >= 80 && score <= 90) && (age > 20)", true},
	}

	for _, tc := range tests {
		val, err := EvalBool(tc.expr, env)
		if err != nil {
			t.Fatalf("expression '%s' failed: %v", tc.expr, err)
		}
		if val != tc.expected {
			t.Fatalf("expression '%s': expected %v, got %v", tc.expr, tc.expected, val)
		}
	}
}

func TestExpr_BuiltinFunctions(t *testing.T) {
	tests := []struct {
		expr     string
		expected float64
	}{
		{"min(10, 20, 5, 40)", 5},
		{"max(10, 20, 5, 40)", 40},
		{"abs(-12.5)", 12.5},
		{"round(14.6)", 15},
		{"round(14.4)", 14},
	}

	for _, tc := range tests {
		val, err := EvalFloat(tc.expr, nil)
		if err != nil {
			t.Fatalf("function expr '%s' failed: %v", tc.expr, err)
		}
		if val != tc.expected {
			t.Fatalf("function expr '%s': expected %f, got %f", tc.expr, tc.expected, val)
		}
	}
}

func TestExpr_DivisionByZero(t *testing.T) {
	_, err := EvalFloat("100 / 0", nil)
	if err == nil {
		t.Fatalf("expected error on division by zero")
	}

	_, err = EvalFloat("100 % 0", nil)
	if err == nil {
		t.Fatalf("expected error on modulo by zero")
	}
}

func TestExpr_SyntaxErrors(t *testing.T) {
	invalidExprs := []string{
		"10 + ",
		"(10 + 20",
		"10 & 20",
		"min()",
	}

	for _, expr := range invalidExprs {
		if _, err := Eval(expr, nil); err == nil {
			t.Fatalf("expected syntax error for '%s'", expr)
		}
	}
}

func TestExpr_CompiledProgramConcurrent(t *testing.T) {
	prog, err := Compile("price * (1 - discount) > 100")
	if err != nil {
		t.Fatalf("failed to compile program: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			env := map[string]any{
				"price":    150 + idx,
				"discount": 0.1,
			}
			res, err := prog.EvalBool(env)
			if err != nil || !res {
				t.Errorf("expected evaluation to succeed and be true")
			}
		}(i)
	}
	wg.Wait()
}
