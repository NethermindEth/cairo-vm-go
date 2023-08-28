package hintrunner

import "fmt"

// HintRunnerError represents error ocurring during the hint runner
type HintRunnerError struct {
	err error
}

func NewHintRunnerError(err error) *HintRunnerError {
	return &HintRunnerError{err}
}

func (e *HintRunnerError) Error() string {
	return fmt.Sprintf("error in hint runner: %s", e.err.Error())
}

func (e *HintRunnerError) Unwrap() error {
	return e.err
}

// HintError stores information about the hint being executed as well as its cause
type HintError struct {
	hintName string
	err      error
}

func NewHintError(hintName string, err error) *HintError {
	return &HintError{hintName, err}
}

func (e *HintError) Error() string {
	return fmt.Sprintf("error executing hint %s: %s", e.hintName, e.err.Error())
}

func (e *HintError) Unwrap() error {
	return e.err
}

// todo(rodro): Should add custom error for operand?
