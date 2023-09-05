package runner

import "fmt"

type RunnerError struct {
	err error
}

func NewRunnerError(err error) *RunnerError {
	return &RunnerError{err}
}

func (e *RunnerError) Error() string {
	return fmt.Sprintf("runner error: %s", e.err.Error())
}

func (e *RunnerError) Unwrap() error {
	return e.err
}
