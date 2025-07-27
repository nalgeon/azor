package azor

import (
	"context"
	"fmt"

	"github.com/nalgeon/azor/promise"
)

// Promise represents the result of an asynchronous call
// that will be available later. The result can be
// either a value or an error.
//
// Promise is a simple type-safe wrapper for [promise.Promise].
// It only runs the given function asynchronously and returns the result.
// Other features like Then or Catch are not supported.
//
// Do not create promises directly, use [Run] instead.
type Promise[T any] struct {
	p *promise.Promise
}

// Run calls the given function asynchronously and returns a [Promise].
// The promise will resolve with the function's result,
// or reject with an error if the function returns one or panics.
//
// Panics if the given function is nil.
func Run[T any](fn func() (T, error)) *Promise[T] {
	if fn == nil {
		panic("azor: nil function")
	}
	return &Promise[T]{
		promise.New(func(resolve func(any), reject func(error)) {
			val, err := fn()
			if err != nil {
				reject(err)
				return
			}
			resolve(val)
		}),
	}
}

// Get waits for the promise to settle and returns the result.
// If the context is canceled before the promise is settled,
// returns a zero value and the context's error.
//
// Get is safe to call from multiple goroutines.
func (p *Promise[T]) Get(ctx context.Context) (T, error) {
	var pval T
	var perr error

	if ctx == nil {
		ctx = context.Background()
	}

	// When the promise settles, collect the result.
	np := p.p.Then(func(value any) any {
		val, ok := value.(T)
		if ok {
			pval = val
		} else {
			// This should never happen given the Run design,
			// which only accepts functions that return T.
			panic(fmt.Sprintf("azor: got value type %T, want %T", value, pval))
		}
		return nil
	}, func(err error) any {
		perr = err
		return nil
	})

	// Wait for the promise to settle
	// or the context to cancel.
	select {
	case <-np.Done():
		return pval, perr
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// Done returns a channel that is closed when
// the promise is settled.
func (p *Promise[T]) Done() <-chan struct{} {
	return p.p.Done()
}
