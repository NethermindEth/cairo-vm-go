package utils

import "fmt"

type SafeMathError struct {
	msg string
}

func NewSafeOffsetError(a uint64, b int16) *SafeMathError {
	return &SafeMathError{
		msg: fmt.Sprintf("offset calculation of %d using %d is out of [0, 2**64) range", a, b),
	}
}

func (e *SafeMathError) Error() string {
	return fmt.Sprintf("math error: %s", e.msg)
}

func (e *SafeMathError) Unwrap() error {
	return nil
}
