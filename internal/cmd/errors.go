package cmd

import "fmt"

type ErrWithExitCode struct {
	exitCode int
	err      error
}

func NewErrWithExitCode(originalError error, exitCode int) *ErrWithExitCode {
	return &ErrWithExitCode{
		exitCode: exitCode,
		err:      originalError,
	}
}

func (e *ErrWithExitCode) Unwrap() error {
	return e.err
}

func (e *ErrWithExitCode) Error() string {
	if e.err == nil {
		return fmt.Sprintf("ErrWithExitCode: %d", e.exitCode)
	}

	return e.err.Error()
}
