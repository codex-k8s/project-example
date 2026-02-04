package shutdown

import (
	"context"
	"errors"
)

// Closer is a shutdown hook that must be safe to call during graceful termination.
type Closer func(context.Context) error

// Run executes closers in reverse order (LIFO) and joins errors via errors.Join.
func Run(ctx context.Context, closers ...Closer) error {
	var errs []error
	for i := len(closers) - 1; i >= 0; i-- {
		if closers[i] == nil {
			continue
		}
		if err := closers[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
