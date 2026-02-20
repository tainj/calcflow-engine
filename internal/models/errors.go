package models

import "errors"

var (
	ErrDivisionByZero       = errors.New("division by zero")
	ErrNonExistingOperation = errors.New("operation does not exist or not implemented")
	ErrCovertExample        = errors.New("line is not a mathematical expression or contains an error")
)
