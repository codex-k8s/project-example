package errs

import "fmt"

// Validation indicates a request or business rule validation error.
type Validation struct {
	Field string
	Msg   string
}

// Error implements error.
func (e Validation) Error() string {
	if e.Field == "" {
		return "validation error"
	}
	if e.Msg == "" {
		return fmt.Sprintf("validation error: %s", e.Field)
	}
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Msg)
}

// Unauthorized indicates missing/invalid authentication.
type Unauthorized struct{ Msg string }

// Error implements error.
func (e Unauthorized) Error() string {
	if e.Msg == "" {
		return "unauthorized"
	}
	return "unauthorized: " + e.Msg
}

// Forbidden indicates lack of permission for an action.
type Forbidden struct{ Msg string }

// Error implements error.
func (e Forbidden) Error() string {
	if e.Msg == "" {
		return "forbidden"
	}
	return "forbidden: " + e.Msg
}

// NotFound indicates a missing resource.
type NotFound struct{ Msg string }

// Error implements error.
func (e NotFound) Error() string {
	if e.Msg == "" {
		return "not found"
	}
	return "not found: " + e.Msg
}

// Conflict indicates a state conflict (e.g. already exists, precondition failed).
type Conflict struct{ Msg string }

// Error implements error.
func (e Conflict) Error() string {
	if e.Msg == "" {
		return "conflict"
	}
	return "conflict: " + e.Msg
}
