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

// OperandError is returned when the error is detected during an operand get/resolve execution
type OperandError struct {
	operandName string
	err         error
}

func NewOperandError(operandName string, err error) *OperandError {
	return &OperandError{operandName, err}
}

func (e *OperandError) Error() string {
	return fmt.Sprintf("failed to get/resolve operand %s: %s", e.operandName, e.err.Error())
}

func (e *OperandError) Unwrap() error {
	return e.err
}
