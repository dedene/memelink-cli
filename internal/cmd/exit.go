// Package cmd implements the memelink CLI commands and Kong parser setup.
package cmd

import "errors"

// ExitError wraps an error with a process exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil || e.Err == nil {
		return "exit"
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// ExitCode extracts the exit code from an error.
// Returns 0 for nil, the embedded code for ExitError, 1 otherwise.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *ExitError
	if errors.As(err, &ee) && ee != nil {
		if ee.Code < 0 {
			return 1
		}
		return ee.Code
	}
	return 1
}

// exitPanic is used by the kong.Exit trick to intercept os.Exit calls.
type exitPanic struct{ code int }
