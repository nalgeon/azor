// Package azor provides promises and async/await functionality.
package azor

import (
	"context"
)

// AsyncFunc is a function that runs asynchronously
// and returns a [Promise] when called.
type AsyncFunc[T any] func() *Promise[T]

// Async creates an asynchronous function from the given function.
// Panics if the function is nil.
func Async[T any](fn func() (T, error)) AsyncFunc[T] {
	if fn == nil {
		panic("azor: nil function")
	}
	return func() *Promise[T] {
		return Run(fn)
	}
}

// Await waits for the promise to settle and returns its result.
// If the context is canceled before the promise is settled,
// returns a zero value and the context's error.
// Panics if the promise is nil.
func Await[T any](ctx context.Context, p *Promise[T]) (T, error) {
	if p == nil {
		panic("azor: nil promise")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return p.Get(ctx)
}
