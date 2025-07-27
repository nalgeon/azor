package promise

import (
	"errors"
	"sync"
	"testing"
	"time"
)

var errDummy = errors.New("dummy")
var dummy = struct{ dummy string }{"dummy"}

// Not applicable:
// 2.2.1: Both `onFulfilled` and `onRejected` are optional arguments.
// 2.2.1.1: If `onFulfilled` is not a function, it must be ignored.
// 2.2.1.2: If `onRejected` is not a function, it must be ignored.
// 2.3.3.1: Let `then` be `x.then`.
// 2.3.3.2: If retrieving the property `x.then` results in a thrown exception `e`, reject `promise` with `e` as the reason.

// Not implemented:
// 2.2.6.1: If/when promise is fulfilled, all respective onFulfilled callbacks must execute in the order of their originating calls to then.
// 2.2.6.2: If/when `promise` is rejected, all respective `onRejected` callbacks must execute in the order of their originating calls to `then`.
// 2.3.3.2: If retrieving the property `x.then` results in a thrown exception `e`, reject `promise` with `e` as the reason.
// 2.3.3.3: If `then` is a function, call it with `x` as `this`, first argument `resolvePromise`, and second argument `rejectPromise`.

// testFulfilled tests the behavior of a promise when it is fulfilled.
// It runs the provided test function with 3 cases:
//   - a promise that is already fulfilled,
//   - a promise that is fulfilled immediately,
//   - a promise that is fulfilled after a delay.
func testFulfilled(t *testing.T, value any, test func(t *testing.T, p *Promise, wg *sync.WaitGroup)) {
	t.Run("already fulfilled", func(t *testing.T) {
		var wg sync.WaitGroup
		test(t, Resolve(value), &wg)
		wg.Wait()
	})
	t.Run("fulfill immediately", func(t *testing.T) {
		var wg sync.WaitGroup
		p := newPromise()
		test(t, p, &wg)
		p.resolve(value)
		wg.Wait()
	})
	t.Run("fulfill delayed", func(t *testing.T) {
		var wg sync.WaitGroup
		p := newPromise()
		test(t, p, &wg)
		go func() {
			time.Sleep(time.Millisecond)
			p.resolve(value)
		}()
		wg.Wait()
	})
}

// testRejected tests the behavior of a promise when it is rejected.
// It runs the provided test function with 3 cases:
//   - a promise that is already rejected,
//   - a promise that is rejected immediately,
//   - a promise that is rejected after a delay.
func testRejected(t *testing.T, err error, test func(t *testing.T, p *Promise, wg *sync.WaitGroup)) {
	t.Run("already rejected", func(t *testing.T) {
		var wg sync.WaitGroup
		test(t, Reject(err), &wg)
		wg.Wait()
	})
	t.Run("reject immediately", func(t *testing.T) {
		var wg sync.WaitGroup
		p := newPromise()
		test(t, p, &wg)
		p.reject(err)
		wg.Wait()
	})
	t.Run("reject delayed", func(t *testing.T) {
		var wg sync.WaitGroup
		p := newPromise()
		test(t, p, &wg)
		go func() {
			time.Sleep(time.Millisecond)
			p.reject(err)
		}()
		wg.Wait()
	})
}

// testResolution tests the behavior of a promise when it is
// resolved or rejected with a value created by newX function.
func testResolution(t *testing.T, newX func() *Promise, test func(t *testing.T, p *Promise, wg *sync.WaitGroup)) {
	t.Run("from fulfilled", func(t *testing.T) {
		var wg sync.WaitGroup
		p := Resolve(dummy).Then(func(value any) any {
			return newX()
		})
		test(t, p, &wg)
		wg.Wait()
	})
	t.Run("from rejected", func(t *testing.T) {
		var wg sync.WaitGroup
		p := Reject(errDummy).Then(nil, func(err error) any {
			return newX()
		})
		test(t, p, &wg)
		wg.Wait()
	})
}
