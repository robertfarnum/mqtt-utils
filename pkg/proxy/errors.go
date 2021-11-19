package proxy

import (
	"strings"
	"sync"
)

// Errors is the set of errors returned by set of goroutines
type Errors struct {
	errs []error
	mu   sync.Mutex
}

// Add an error to Errors thread safe
func (e *Errors) Add(err error) {
	e.mu.Lock()

	e.errs = append(e.errs, err)

	e.mu.Unlock()
}

// Error returns the string format for the Errors thread safe
func (e *Errors) Error() string {
	e.mu.Lock()

	acc := make([]string, 0, len(e.errs))
	for _, ce := range e.errs {
		if ce != nil {
			acc = append(acc, ce.Error())
		}
	}

	e.mu.Unlock()

	if 0 < len(acc) {
		return strings.Join(acc, ",")
	}

	return "N/A"
}
