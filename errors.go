package cli

import (
	"errors"
	"strings"
)

var (
	errHelpRequested    = errors.New("help requested")
	errVersionRequested = errors.New("version requested")
)

type MultiError struct {
	Errors []error
}

func NewMultiError(err ...error) MultiError {
	return MultiError{Errors: err}
}

func (m MultiError) Error() string {
	errs := make([]string, len(m.Errors))
	for i, err := range m.Errors {
		errs[i] = err.Error()
	}

	return strings.Join(errs, "\n")
}