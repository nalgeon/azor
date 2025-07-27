package promise

import (
	"errors"
	"sync"
	"testing"
)

func TestThenChain(t *testing.T) {
	t.Log("2.2.7: `then` must return a promise: `promise2 = promise1.then(onFulfilled, onRejected)`")
	t.Run("handler error", func(t *testing.T) {
		t.Log("2.2.7.2: If either `onFulfilled` or `onRejected` throws an exception `e`, `promise2` must be rejected with `e` as the reason.")
		t.Run("onFulfilled error", func(t *testing.T) {
			testFulfilled(t, dummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(1)
				p2 := p.Then(func(value any) any {
					return errDummy
				})

				p2.Then(func(value any) any {
					t.Error("onFulfilled should not be called")
					return nil
				}, func(err error) any {
					if !errors.Is(err, errDummy) {
						t.Errorf("got %v, want %v", err, errDummy)
					}
					wg.Done()
					return nil
				})
			})
		})
		t.Run("onRejected error", func(t *testing.T) {
			testRejected(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(1)
				errSentinel := errors.New("sentinel error")

				p2 := p.Then(nil, func(err error) any {
					return errSentinel
				})

				p2.Then(func(value any) any {
					t.Error("onFulfilled should not be called")
					return nil
				}, func(err error) any {
					if !errors.Is(err, errSentinel) {
						t.Errorf("got %v, want %v", err, errSentinel)
					}
					wg.Done()
					return nil
				})
			})
		})
	})
	t.Run("nil onFulfilled", func(t *testing.T) {
		t.Log("2.2.7.3: If `onFulfilled` is not a function and `promise1` is fulfilled, `promise2` must be fulfilled with the same value.")
		testFulfilled(t, dummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p2 := p.Then(nil)
			p2.Then(func(value any) any {
				if value != dummy {
					t.Errorf("got %v, want %v", value, dummy)
				}
				wg.Done()
				return nil
			})
		})

	})
	t.Run("nil onRejected", func(t *testing.T) {
		testRejected(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p2 := p.Then(nil, nil)
			p2.Then(nil, func(err error) any {
				if !errors.Is(err, errDummy) {
					t.Errorf("got %v, want %v", err, errDummy)
				}
				wg.Done()
				return nil
			})
		})
	})
}
