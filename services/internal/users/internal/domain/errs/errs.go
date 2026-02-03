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

type NotFound struct {
	Entity string
	ID     any
}

func (e NotFound) Error() string {
	if e.Entity == "" {
		return "not found"
	}
	if e.ID == nil {
		return fmt.Sprintf("%s not found", e.Entity)
	}
	return fmt.Sprintf("%s not found: %v", e.Entity, e.ID)
}

type Unauthorized struct{ Msg string }

func (e Unauthorized) Error() string {
	if e.Msg == "" {
		return "unauthorized"
	}
	return "unauthorized: " + e.Msg
}

type Conflict struct{ Msg string }

func (e Conflict) Error() string {
	if e.Msg == "" {
		return "conflict"
	}
	return "conflict: " + e.Msg
}
