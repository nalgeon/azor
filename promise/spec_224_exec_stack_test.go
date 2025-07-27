package promise

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestExecStack(t *testing.T) {
	t.Log("2.2.4: `onFulfilled` or `onRejected` must not be called until the execution context stack contains only platform code.")
	t.Run("then returns before promise settled", func(t *testing.T) {
		testFulfilled(t, dummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			var thenReturned atomic.Bool

			wg.Add(1)
			p.Then(func(value any) any {
				if !thenReturned.Load() {
					t.Errorf("then must return before promise settled")
				}
				wg.Done()
				return nil
			})

			thenReturned.Store(true)
		})
		testRejected(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			var thenReturned atomic.Bool

			wg.Add(1)
			p.Then(nil, func(err error) any {
				if !thenReturned.Load() {
					t.Errorf("then must return before promise settled")
				}
				wg.Done()
				return nil
			})

			thenReturned.Store(true)
		})
	})
	t.Run("clean stack fulfillment", func(t *testing.T) {
		t.Run("onFulfilled added before promise fulfilled", func(t *testing.T) {
			var onFulfilledCalled atomic.Bool

			p := newPromise()
			p.Then(func(value any) any {
				onFulfilledCalled.Store(true)
				return nil
			})

			p.resolve(dummy)

			if onFulfilledCalled.Load() {
				t.Error("onFulfilled should not be called yet")
			}
		})
		t.Run("onFulfilled added after promise fulfilled", func(t *testing.T) {
			var onFulfilledCalled atomic.Bool

			p := newPromise()
			p.resolve(dummy)

			p.Then(func(value any) any {
				onFulfilledCalled.Store(true)
				return nil
			})

			if onFulfilledCalled.Load() {
				t.Error("onFulfilled should not be called yet")
			}
		})
		t.Run("onFulfilled added inside onFulfilled", func(t *testing.T) {
			done := make(chan struct{})
			var firstOnFulfilledFinished atomic.Bool

			p := Resolve(dummy)
			p.Then(func(value any) any {
				p.Then(func(value any) any {
					if !firstOnFulfilledFinished.Load() {
						t.Error("first onFulfilled should have finished")
					}
					close(done)
					return nil
				})
				firstOnFulfilledFinished.Store(true)
				return nil
			})

			<-done
		})
		t.Run("onFulfilled added inside onRejected", func(t *testing.T) {
			done := make(chan struct{})
			var firstOnRejectedFinished atomic.Bool

			p1 := Reject(errDummy)
			p2 := Resolve(dummy)

			p1.Then(nil, func(err error) any {
				p2.Then(func(value any) any {
					if !firstOnRejectedFinished.Load() {
						t.Error("first onRejected should have finished")
					}
					close(done)
					return nil
				})
				firstOnRejectedFinished.Store(true)
				return nil
			})

			<-done
		})
	})
	t.Run("clean stack rejection", func(t *testing.T) {
		t.Run("onRejected added before promise rejected", func(t *testing.T) {
			var onRejectedCalled atomic.Bool

			p := newPromise()
			p.Then(nil, func(err error) any {
				onRejectedCalled.Store(true)
				return nil
			})

			p.reject(errDummy)

			if onRejectedCalled.Load() {
				t.Error("onRejected should not be called yet")
			}
		})
		t.Run("onRejected added after promise rejected", func(t *testing.T) {
			var onRejectedCalled atomic.Bool

			p := newPromise()
			p.reject(errDummy)

			p.Then(nil, func(err error) any {
				onRejectedCalled.Store(true)
				return nil
			})

			if onRejectedCalled.Load() {
				t.Error("onFulfilled should not be called yet")
			}
		})
		t.Run("onRejected added inside onFulfilled", func(t *testing.T) {
			done := make(chan struct{})
			var firstOnFulfilledFinished atomic.Bool

			p1 := Resolve(dummy)
			p2 := Reject(errDummy)

			p1.Then(func(value any) any {
				p2.Then(nil, func(err error) any {
					if !firstOnFulfilledFinished.Load() {
						t.Error("first onFulfilled should have finished")
					}
					close(done)
					return nil
				})
				firstOnFulfilledFinished.Store(true)
				return nil
			})

			<-done
		})
		t.Run("onRejected added inside onRejected", func(t *testing.T) {
			done := make(chan struct{})
			var firstOnRejectedFinished atomic.Bool

			p := Reject(errDummy)
			p.Then(nil, func(err error) any {
				p.Then(nil, func(err error) any {
					if !firstOnRejectedFinished.Load() {
						t.Error("first onRejected should have finished")
					}
					close(done)
					return nil
				})
				firstOnRejectedFinished.Store(true)
				return nil
			})

			<-done
		})
	})
}
