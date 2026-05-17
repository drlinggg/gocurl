package types

import "fmt"

type ErrValidation struct {
	Field string
}

func (e *ErrValidation) Error() string {
	return e.Field + " is required"
}

type ErrStatusRange struct {
	Code int
}

func (e *ErrStatusRange) Error() string {
	return fmt.Sprintf("status %d out of range [0, 999]", e.Code)
}

type ErrInvalidColor struct {
	Value string
}

func (e *ErrInvalidColor) Error() string {
	return fmt.Sprintf("invalid color %q: expected 6-digit hex", e.Value)
}
