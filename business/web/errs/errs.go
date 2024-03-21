// Package errs provides types and support related to web v1 functionality.
package errs

import "errors"

// Response is the form used for API responses from failures in the API.
type Response struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// Trusted is used to pass an error during the request through the
// application with web specific context.
type Trusted struct {
	Err    error
	Status int
}

// NewTrusted wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewTrusted(err error, status int) error {
	return &Trusted{err, status}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (re *Trusted) Error() string {
	return re.Err.Error()
}

// IsTrusted checks if an error of type RequestError exists.
func IsTrusted(err error) bool {
	var re *Trusted
	return errors.As(err, &re)
}

// GetTrusted returns a copy of the RequestError pointer.
func GetTrusted(err error) *Trusted {
	var re *Trusted
	if !errors.As(err, &re) {
		return nil
	}
	return re
}
