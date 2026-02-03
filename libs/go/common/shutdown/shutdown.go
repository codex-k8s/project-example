package shutdown

import (
	"context"
	"errors"
)

type Closer func(context.Context) error

// Run выполняет close-функции в обратном порядке (LIFO), собирая ошибки в errors.Join.
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
