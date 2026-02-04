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

// NotFound indicates a missing domain entity.
type NotFound struct {
	Entity string
	ID     any
}

// Error implements error.
func (e NotFound) Error() string {
	if e.Entity == "" {
		return "not found"
	}
	if e.ID == nil {
		return fmt.Sprintf("%s not found", e.Entity)
	}
	return fmt.Sprintf("%s not found: %v", e.Entity, e.ID)
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

// Conflict indicates a state conflict (e.g. already exists, precondition failed).
type Conflict struct{ Msg string }

// Error implements error.
func (e Conflict) Error() string {
	if e.Msg == "" {
		return "conflict"
	}
	return "conflict: " + e.Msg
}
