package valueobjects

import "fmt"

// StatusCode wraps an HTTP-style status code validation.
type StatusCode struct {
	value int
}

// NewStatusCode ensures the status code falls within 100-599 inclusive.
func NewStatusCode(code int) (StatusCode, error) {
	if code < 100 || code > 599 {
		return StatusCode{}, fmt.Errorf("invalid status code: %d", code)
	}
	return StatusCode{value: code}, nil
}

// Int exposes the integer representation.
func (s StatusCode) Int() int {
	return s.value
}

// IsZero reports whether the status code has been set.
func (s StatusCode) IsZero() bool {
	return s.value == 0
}
