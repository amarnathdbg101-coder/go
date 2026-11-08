package expr

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// TokenType classifies lexer tokens.
type TokenType int

const (
	tokEOF TokenType = iota
	tokNumber
	tokString
	tokBool
	tokIdent
	tokPlus
	tokMinus
	tokMul
	tokDiv
	tokMod
	tokEQ
	tokNEQ
	tokLT
	tokLTE
	tokGT
	tokGTE
	tokAnd
	tokOr
	tokNot
	tokLParen
	tokRParen
	tokComma
)

type token struct {
	typ TokenType
	val string
	num float64
	b   bool
}

// Lexer breaks expression strings into a stream of tokens.
type lexer struct {
	input []rune
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: []rune(input)}
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *lexer) next() rune {
	r := l.peek()
	if r != 0 {
		l.pos++
	}
	return r
}

func (l *lexer) skipWhitespace() {
	for unicode.IsSpace(l.peek()) {
		l.next()
	}
}

func (l *lexer) nextToken() (token, error) {
	l.skipWhitespace()
	r := l.peek()
	if r == 0 {
		return token{typ: tokEOF}, nil
	}

	// Single/Double char operators
	switch r {
	case '+':
		l.next()
		return token{typ: tokPlus, val: "+"}, nil
	case '-':
		l.next()
		return token{typ: tokMinus, val: "-"}, nil
	case '*':
		l.next()
		return token{typ: tokMul, val: "*"}, nil
	case '/':
		l.next()
		return token{typ: tokDiv, val: "/"}, nil
	case '%':
		l.next()
		return token{typ: tokMod, val: "%"}, nil
	case '(':
		l.next()
		return token{typ: tokLParen, val: "("}, nil
	case ')':
		l.next()
		return token{typ: tokRParen, val: ")"}, nil
	case ',':
		l.next()
		return token{typ: tokComma, val: ","}, nil
	case '!':
		l.next()
		if l.peek() == '=' {
			l.next()
			return token{typ: tokNEQ, val: "!="}, nil
		}
		return token{typ: tokNot, val: "!"}, nil
	case '=':
		l.next()
		if l.peek() == '=' {
			l.next()
			return token{typ: tokEQ, val: "=="}, nil
		}
		return token{}, fmt.Errorf("%w: unexpected '=' at pos %d, did you mean '=='?", ErrSyntax, l.pos)
	case '<':
		l.next()
		if l.peek() == '=' {
			l.next()
			return token{typ: tokLTE, val: "<="}, nil
		}
		return token{typ: tokLT, val: "<"}, nil
	case '>':
		l.next()
		if l.peek() == '=' {
			l.next()
			return token{typ: tokGTE, val: ">="}, nil
		}
		return token{typ: tokGT, val: ">"}, nil
	case '&':
		l.next()
		if l.peek() == '&' {
			l.next()
			return token{typ: tokAnd, val: "&&"}, nil
		}
		return token{}, fmt.Errorf("%w: single '&' at pos %d", ErrSyntax, l.pos)
	case '|':
		l.next()
		if l.peek() == '|' {
			l.next()
			return token{typ: tokOr, val: "||"}, nil
		}
		return token{}, fmt.Errorf("%w: single '|' at pos %d", ErrSyntax, l.pos)
	case '"', '\'':
		quote := l.next()
		var sb strings.Builder
		for {
			ch := l.next()
			if ch == 0 {
				return token{}, fmt.Errorf("%w: unclosed string literal", ErrSyntax)
			}
			if ch == quote {
				break
			}
			sb.WriteRune(ch)
		}
		return token{typ: tokString, val: sb.String()}, nil
	}

	// Numbers
	if unicode.IsDigit(r) || (r == '.' && l.pos+1 < len(l.input) && unicode.IsDigit(l.input[l.pos+1])) {
		var sb strings.Builder
		for unicode.IsDigit(l.peek()) || l.peek() == '.' {
			sb.WriteRune(l.next())
		}
		valStr := sb.String()
		num, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return token{}, fmt.Errorf("%w: invalid number '%s'", ErrSyntax, valStr)
		}
		return token{typ: tokNumber, val: valStr, num: num}, nil
	}

	// Identifiers or boolean keywords
	if unicode.IsLetter(r) || r == '_' {
		var sb strings.Builder
		for unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_' {
			sb.WriteRune(l.next())
		}
		ident := sb.String()
		switch strings.ToLower(ident) {
		case "true":
			return token{typ: tokBool, val: "true", b: true}, nil
		case "false":
			return token{typ: tokBool, val: "false", b: false}, nil
		default:
			return token{typ: tokIdent, val: ident}, nil
		}
	}

	return token{}, fmt.Errorf("%w: unexpected character '%c' at position %d", ErrSyntax, r, l.pos)
}

// AST Nodes
type node interface {
	eval(env map[string]any) (any, error)
}

type numberNode struct{ val float64 }

func (n *numberNode) eval(env map[string]any) (any, error) { return n.val, nil }

type stringNode struct{ val string }

func (n *stringNode) eval(env map[string]any) (any, error) { return n.val, nil }

type boolNode struct{ val bool }

func (n *boolNode) eval(env map[string]any) (any, error) { return n.val, nil }

type identNode struct{ name string }

func (n *identNode) eval(env map[string]any) (any, error) {
	if val, ok := env[n.name]; ok {
		return toFloatOrVal(val), nil
	}
	return nil, fmt.Errorf("%w: '%s'", ErrUnknownIdentifier, n.name)
}

func toFloatOrVal(v any) any {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return v
	}
}

type unaryNode struct {
	op   TokenType
	expr node
}

func (n *unaryNode) eval(env map[string]any) (any, error) {
	val, err := n.expr.eval(env)
	if err != nil {
		return nil, err
	}

	switch n.op {
	case tokMinus:
		num, ok := val.(float64)
		if !ok {
			return nil, fmt.Errorf("%w: unary '-' requires numeric operand", ErrTypeMismatch)
		}
		return -num, nil
	case tokNot:
		b, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: '!' requires boolean operand", ErrTypeMismatch)
		}
		return !b, nil
	default:
		return nil, ErrUnexpectedToken
	}
}

type binaryNode struct {
	op    TokenType
	left  node
	right node
}

func (n *binaryNode) eval(env map[string]any) (any, error) {
	leftVal, err := n.left.eval(env)
	if err != nil {
		return nil, err
	}

	// Short-circuit logical operations
	if n.op == tokAnd {
		lb, ok := leftVal.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: '&&' requires boolean", ErrTypeMismatch)
		}
		if !lb {
			return false, nil
		}
		rightVal, err := n.right.eval(env)
		if err != nil {
			return nil, err
		}
		rb, ok := rightVal.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: '&&' requires boolean", ErrTypeMismatch)
		}
		return rb, nil
	}

	if n.op == tokOr {
		lb, ok := leftVal.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: '||' requires boolean", ErrTypeMismatch)
		}
		if lb {
			return true, nil
		}
		rightVal, err := n.right.eval(env)
		if err != nil {
			return nil, err
		}
		rb, ok := rightVal.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: '||' requires boolean", ErrTypeMismatch)
		}
		return rb, nil
	}

	rightVal, err := n.right.eval(env)
	if err != nil {
		return nil, err
	}

	switch n.op {
	case tokPlus:
		// Numeric addition or string concatenation
		if ln, lok := leftVal.(float64); lok {
			if rn, rok := rightVal.(float64); rok {
				return ln + rn, nil
			}
		}
		if ls, lok := leftVal.(string); lok {
			return ls + fmt.Sprint(rightVal), nil
		}
		return nil, fmt.Errorf("%w: '+' cannot operate on given types", ErrTypeMismatch)

	case tokMinus:
		ln, lok := leftVal.(float64)
		rn, rok := rightVal.(float64)
		if !lok || !rok {
			return nil, fmt.Errorf("%w: '-' requires numbers", ErrTypeMismatch)
		}
		return ln - rn, nil

	case tokMul:
		ln, lok := leftVal.(float64)
		rn, rok := rightVal.(float64)
		if !lok || !rok {
			return nil, fmt.Errorf("%w: '*' requires numbers", ErrTypeMismatch)
		}
		return ln * rn, nil

	case tokDiv:
		ln, lok := leftVal.(float64)
		rn, rok := rightVal.(float64)
		if !lok || !rok {
			return nil, fmt.Errorf("%w: '/' requires numbers", ErrTypeMismatch)
		}
		if rn == 0 {
			return nil, ErrDivisionByZero
		}
		return ln / rn, nil

	case tokMod:
		ln, lok := leftVal.(float64)
		rn, rok := rightVal.(float64)
		if !lok || !rok {
			return nil, fmt.Errorf("%w: '%%' requires numbers", ErrTypeMismatch)
		}
		if rn == 0 {
			return nil, ErrDivisionByZero
		}
		return math.Mod(ln, rn), nil

	case tokEQ:
		return leftVal == rightVal, nil

	case tokNEQ:
		return leftVal != rightVal, nil

	case tokLT:
		ln, lok := leftVal.(float64)
		rn, rok := rightVal.(float64)
		if lok && rok {
			return ln < rn, nil
		}
		if ls, lok := leftVal.(string); lok {
			if rs, rok := rightVal.(string); rok {
				return ls < rs, nil
			}
		}
		return nil, fmt.Errorf("%w: '<' requires comparable types", ErrTypeMismatch)

	case tokLTE:
		ln, lok := leftVal.(float64)
		rn, rok := rightVal.(float64)
		if lok && rok {
			return ln <= rn, nil
		}
		return nil, fmt.Errorf("%w: '<=' requires numbers", ErrTypeMismatch)

	case tokGT:
		ln, lok := leftVal.(float64)
		rn, rok := rightVal.(float64)
		if lok && rok {
			return ln > rn, nil
		}
		if ls, lok := leftVal.(string); lok {
			if rs, rok := rightVal.(string); rok {
				return ls > rs, nil
			}
		}
		return nil, fmt.Errorf("%w: '>' requires comparable types", ErrTypeMismatch)

	case tokGTE:
		ln, lok := leftVal.(float64)
		rn, rok := rightVal.(float64)
		if lok && rok {
			return ln >= rn, nil
		}
		return nil, fmt.Errorf("%w: '>=' requires numbers", ErrTypeMismatch)
	}

	return nil, ErrUnexpectedToken
}

type callNode struct {
	fn   string
	args []node
}

func (n *callNode) eval(env map[string]any) (any, error) {
	var evaluated []float64
	for _, arg := range n.args {
		v, err := arg.eval(env)
		if err != nil {
			return nil, err
		}
		num, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("%w: built-in math function '%s' expects numeric arguments", ErrTypeMismatch, n.fn)
		}
		evaluated = append(evaluated, num)
	}

	switch strings.ToLower(n.fn) {
	case "min":
		if len(evaluated) < 2 {
			return nil, fmt.Errorf("%w: min() requires at least 2 arguments", ErrSyntax)
		}
		m := evaluated[0]
		for _, v := range evaluated[1:] {
			if v < m {
				m = v
			}
		}
		return m, nil
	case "max":
		if len(evaluated) < 2 {
			return nil, fmt.Errorf("%w: max() requires at least 2 arguments", ErrSyntax)
		}
		m := evaluated[0]
		for _, v := range evaluated[1:] {
			if v > m {
				m = v
			}
		}
		return m, nil
	case "abs":
		if len(evaluated) != 1 {
			return nil, fmt.Errorf("%w: abs() takes exactly 1 argument", ErrSyntax)
		}
		return math.Abs(evaluated[0]), nil
	case "round":
		if len(evaluated) != 1 {
			return nil, fmt.Errorf("%w: round() takes exactly 1 argument", ErrSyntax)
		}
		return math.Round(evaluated[0]), nil
	default:
		return nil, fmt.Errorf("%w: unknown function '%s'", ErrSyntax, n.fn)
	}
}

// Parser parses tokens into an AST using recursive descent.
type parser struct {
	tokens []token
	pos    int
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{typ: tokEOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) next() token {
	tok := p.peek()
	if tok.typ != tokEOF {
		p.pos++
	}
	return tok
}

func (p *parser) match(typ TokenType) bool {
	if p.peek().typ == typ {
		p.next()
		return true
	}
	return false
}

func (p *parser) parseExpression() (node, error) {
	return p.parseOr()
}

func (p *parser) parseOr() (node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.match(tokOr) {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: tokOr, left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (node, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.match(tokAnd) {
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: tokAnd, left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseEquality() (node, error) {
	left, err := p.parseRelational()
	if err != nil {
		return nil, err
	}
	for {
		if p.match(tokEQ) {
			right, err := p.parseRelational()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokEQ, left: left, right: right}
		} else if p.match(tokNEQ) {
			right, err := p.parseRelational()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokNEQ, left: left, right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseRelational() (node, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	for {
		if p.match(tokLT) {
			right, err := p.parseAdditive()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokLT, left: left, right: right}
		} else if p.match(tokLTE) {
			right, err := p.parseAdditive()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokLTE, left: left, right: right}
		} else if p.match(tokGT) {
			right, err := p.parseAdditive()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokGT, left: left, right: right}
		} else if p.match(tokGTE) {
			right, err := p.parseAdditive()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokGTE, left: left, right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseAdditive() (node, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	for {
		if p.match(tokPlus) {
			right, err := p.parseMultiplicative()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokPlus, left: left, right: right}
		} else if p.match(tokMinus) {
			right, err := p.parseMultiplicative()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokMinus, left: left, right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseMultiplicative() (node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		if p.match(tokMul) {
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokMul, left: left, right: right}
		} else if p.match(tokDiv) {
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokDiv, left: left, right: right}
		} else if p.match(tokMod) {
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: tokMod, left: left, right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseUnary() (node, error) {
	if p.match(tokMinus) {
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNode{op: tokMinus, expr: expr}, nil
	}
	if p.match(tokNot) {
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNode{op: tokNot, expr: expr}, nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (node, error) {
	tok := p.next()
	switch tok.typ {
	case tokNumber:
		return &numberNode{val: tok.num}, nil
	case tokString:
		return &stringNode{val: tok.val}, nil
	case tokBool:
		return &boolNode{val: tok.b}, nil
	case tokIdent:
		// Check if function call
		if p.match(tokLParen) {
			var args []node
			if !p.match(tokRParen) {
				for {
					arg, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
					if p.match(tokComma) {
						continue
					}
					if p.match(tokRParen) {
						break
					}
					return nil, fmt.Errorf("%w: expected ',' or ')' in argument list", ErrSyntax)
				}
			}
			return &callNode{fn: tok.val, args: args}, nil
		}
		return &identNode{name: tok.val}, nil
	case tokLParen:
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.match(tokRParen) {
			return nil, fmt.Errorf("%w: missing closing ')'", ErrSyntax)
		}
		return expr, nil
	default:
		return nil, fmt.Errorf("%w: unexpected '%s'", ErrSyntax, tok.val)
	}
}

// Program is a pre-compiled expression AST ready for high-speed evaluation.
type Program struct {
	root node
}

// Compile parses and compiles an expression string into an executable Program.
func Compile(exprStr string) (*Program, error) {
	lex := newLexer(exprStr)
	var tokens []token
	for {
		tok, err := lex.nextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.typ == tokEOF {
			break
		}
	}

	p := &parser{tokens: tokens}
	root, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.peek().typ != tokEOF {
		return nil, fmt.Errorf("%w: trailing tokens after valid expression", ErrSyntax)
	}

	return &Program{root: root}, nil
}

// Eval evaluates the compiled program with the provided environment variables.
func (p *Program) Eval(env map[string]any) (any, error) {
	if p == nil || p.root == nil {
		return nil, ErrSyntax
	}
	if env == nil {
		env = make(map[string]any)
	}
	return p.root.eval(env)
}

// EvalBool evaluates the program and asserts a boolean result.
func (p *Program) EvalBool(env map[string]any) (bool, error) {
	res, err := p.Eval(env)
	if err != nil {
		return false, err
	}
	b, ok := res.(bool)
	if !ok {
		return false, fmt.Errorf("%w: result is not a boolean (%T)", ErrTypeMismatch, res)
	}
	return b, nil
}

// EvalFloat evaluates the program and asserts a numeric float64 result.
func (p *Program) EvalFloat(env map[string]any) (float64, error) {
	res, err := p.Eval(env)
	if err != nil {
		return 0, err
	}
	f, ok := res.(float64)
	if !ok {
		return 0, fmt.Errorf("%w: result is not a number (%T)", ErrTypeMismatch, res)
	}
	return f, nil
}

// Eval directly compiles and evaluates an expression string.
func Eval(exprStr string, env map[string]any) (any, error) {
	prog, err := Compile(exprStr)
	if err != nil {
		return nil, err
	}
	return prog.Eval(env)
}

// EvalBool directly compiles and evaluates an expression to a boolean.
func EvalBool(exprStr string, env map[string]any) (bool, error) {
	prog, err := Compile(exprStr)
	if err != nil {
		return false, err
	}
	return prog.EvalBool(env)
}

// EvalFloat directly compiles and evaluates an expression to a float64.
func EvalFloat(exprStr string, env map[string]any) (float64, error) {
	prog, err := Compile(exprStr)
	if err != nil {
		return 0, err
	}
	return prog.EvalFloat(env)
}
