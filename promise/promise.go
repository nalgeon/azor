// Package promise provides an implementation
// that's mostly compatible with Promises/A+.
//
// Differences from the spec:
//  1. If you call Then multiple times on the same promise,
//     the handlers may run in any order.
//  2. There is no special handling for "thenables" (objects with a "then" method).
//
// Returning an error from a handler or resolving with an error
// will reject the promise, similar to throwing in JavaScript promises.
package promise

import (
	"errors"
	"fmt"
	"sync"
)

// result represents the result of a promise.
type result struct {
	val any
	err error
}

// Promise represents the eventual completion (or failure)
// of an asynchronous operation and its resulting value.
//
// Promise transitions from pending to settled (fulfilled or rejected).
// Once settled, its result is immutable and all handlers will observe
// the same value or error.
//
// A zero Promise value is unusable. Use [New], [NewContext], [Resolve]
// or [Reject] to create a new promise.
type Promise struct {
	res  result
	done chan struct{}
	once sync.Once
}

// New creates a new promise that will be resolved or rejected
// based on the execution of the given function.
//
// The executor function runs in a new goroutine.
// Panics in the executor are caught and cause the promise to be rejected.
//
// New panics if fn is nil.
func New(fn func(func(any), func(error))) *Promise {
	if fn == nil {
		panic("promise: nil function")
	}
	p := newPromise()
	go func() {
		defer p.rejectOnPanic()
		fn(p.resolve, p.reject)
	}()
	return p
}

// newPromise creates a new pending promise.
func newPromise() *Promise {
	return &Promise{
		done: make(chan struct{}),
	}
}

// Then registers handlers to be called when the promise is fulfilled or rejected.
// Handlers are always executed asynchronously in a new goroutine.
//
// Returns a new promise that will be resolved or rejected based on the results of the handlers.
// The new promise uses the same context as the original promise.
// If the promise is already settled, the handlers are called immediately.
//
// If you call Then multiple times on the same promise, the handlers might run in any order.
// They don't have to run in the order you called Then.
//
// Variadic onRejecteds parameter is a hack to make onRejected optional.
// Only the first onRejected handler is used if multiple are provided.
func (p *Promise) Then(onFulfilled func(any) any, onRejecteds ...func(error) any) *Promise {
	// onFulfilled: if not provided, replace with
	// an identity function (val => val, nil)
	if onFulfilled == nil {
		onFulfilled = func(val any) any { return val }
	}
	// onRejected: if not provided, replace with
	// a thrower function (err => nil, err)
	var onRejected func(error) any
	if len(onRejecteds) > 0 && onRejecteds[0] != nil {
		onRejected = onRejecteds[0]
	} else {
		onRejected = func(err error) any { return err }
	}

	return p.then(onFulfilled, onRejected)
}

// Catch registers a handler to be called when the promise is rejected.
// Returns a new promise that will be resolved or rejected based on the result of the handler.
//
// It's a shorthand for Then(nil, onRejected).
func (p *Promise) Catch(onRejected func(error) any) *Promise {
	return p.Then(nil, onRejected)
}

// Finally registers a handler to be called when the promise is settled (fulfilled or rejected).
// Returns a new promise. If onFinally returns an error or a rejected promise,
// the new promise will reject with that value. Otherwise, the new promise will
// settle with the same state as the current promise.
func (p *Promise) Finally(onFinally func() any) *Promise {
	if onFinally == nil {
		onFinally = func() any { return nil }
	}
	return New(func(resolve func(any), reject func(error)) {
		// Wait for the current promise to settle or be canceled.
		p.wait()

		// Act on the result of the onFinally handler.
		val := onFinally()
		switch x := val.(type) {
		case *Promise:
			// If returned promise if rejected,
			// reject the new promise with its error.
			if x.res.err != nil {
				reject(x.res.err)
				return
			}
		case error:
			// If returned value is an error,
			// reject the new promise with it.
			reject(x)
			return
		}
		// Otherwise, resolve the new promise
		// with the current promise's value.
		if p.res.err != nil {
			reject(p.res.err)
		} else {
			resolve(p.res.val)
		}
	})
}

// Done returns a channel that is closed when
// the promise is settled (fulfilled or rejected).
func (p *Promise) Done() <-chan struct{} {
	return p.done
}

// then returns a new promise that will be resolved or rejected
// based on the results of the onFulfilled/onRejected handlers.
func (p *Promise) then(onFulfilled func(any) any, onRejected func(error) any) *Promise {
	return New(func(resolve func(any), reject func(error)) {
		// Wait for the current promise to settle or be canceled.
		p.wait()

		// Get the value/error from the handlers
		// based on the promise's result.
		var val any
		if p.res.err != nil {
			val = onRejected(p.res.err)
		} else {
			val = onFulfilled(p.res.val)
		}

		if val == p {
			// The promise cannot resolve itself.
			reject(fmt.Errorf("resolve with self: %w", errors.ErrUnsupported))
			return
		}

		// Resolve the new promise according
		// to the value returned by the handler.
		resolve(val)
	})
}

// wait blocks the caller until the promise is settled
// (fulfilled or rejected)
func (p *Promise) wait() {
	<-p.done
}

// resolve resolves the promise with the given value.
// If the value is another promise, recursively waits for it to settle
// and resolves or rejects the current promise with its final value or error.
// If the value is not a promise, it resolves the current promise directly.
func (p *Promise) resolve(value any) {
	switch x := value.(type) {
	case *Promise:
		// If X is a promise, wait for it to settle or cancel.
		x.wait()

		// Resolve or reject the current promise based on the X's result.
		if x.res.err != nil {
			p.reject(x.res.err)
		} else {
			p.resolve(x.res.val)
		}
	case error:
		// If X is an error, reject the current promise.
		p.reject(x)
	default:
		// Otherwise, resolve the current promise with X.
		p.settle(result{val: x})
	}
}

// rejectOnPanic checks if there was a panic during the execution of the promise.
// If there was a panic, it rejects the promise with the panic value.
func (p *Promise) rejectOnPanic() {
	r := recover()
	if r == nil {
		return
	}

	// If there was a panic, reject the promise.
	switch v := r.(type) {
	case error:
		p.reject(v)
	default:
		p.reject(fmt.Errorf("panic: %v", v))
	}
}

// reject rejects the promise with the given error.
func (p *Promise) reject(err error) {
	p.settle(result{err: err})
}

// settle sets the result of the promise
// exactly once in a concurrent-safe manner.
func (p *Promise) settle(res result) {
	p.once.Do(func() {
		p.res = res
		close(p.done)
	})
}

// Resolve resolves a given value to a promise.
// Flattens nested layers of promises into a single
// promise that resolves to the final value.
func Resolve(value any) *Promise {
	p := newPromise()
	p.resolve(value)
	return p
}

// Reject creates a new promise that is
// immediately rejected with the given error.
func Reject(err error) *Promise {
	p := newPromise()
	p.reject(err)
	return p
}
