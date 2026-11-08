package expr

import "errors"

var (
	// ErrSyntax is returned when an expression has invalid syntax or unmatched tokens.
	ErrSyntax = errors.New("expr: syntax error")
	// ErrUnexpectedToken is returned when an unexpected token is encountered during parsing.
	ErrUnexpectedToken = errors.New("expr: unexpected token")
	// ErrUnknownIdentifier is returned when an identifier/variable is missing from the environment.
	ErrUnknownIdentifier = errors.New("expr: unknown identifier")
	// ErrTypeMismatch is returned when an operator is applied to incompatible operand types.
	ErrTypeMismatch = errors.New("expr: type mismatch")
	// ErrDivisionByZero is returned when division or modulo by zero occurs.
	ErrDivisionByZero = errors.New("expr: division by zero")
)
