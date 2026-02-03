package errs

import "fmt"

type Validation struct {
	Field string
	Msg   string
}

func (e Validation) Error() string {
	if e.Field == "" {
		return "validation error"
	}
	if e.Msg == "" {
		return fmt.Sprintf("validation error: %s", e.Field)
	}
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Msg)
}

type Unauthorized struct{ Msg string }

func (e Unauthorized) Error() string {
	if e.Msg == "" {
		return "unauthorized"
	}
	return "unauthorized: " + e.Msg
}

type Forbidden struct{ Msg string }

func (e Forbidden) Error() string {
	if e.Msg == "" {
		return "forbidden"
	}
	return "forbidden: " + e.Msg
}

type NotFound struct{ Msg string }

func (e NotFound) Error() string {
	if e.Msg == "" {
		return "not found"
	}
	return "not found: " + e.Msg
}

type Conflict struct{ Msg string }

func (e Conflict) Error() string {
	if e.Msg == "" {
		return "conflict"
	}
	return "conflict: " + e.Msg
}
