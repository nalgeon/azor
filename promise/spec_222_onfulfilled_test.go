package promise

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestOnFulfilled(t *testing.T) {
	t.Run("fulfillment value", func(t *testing.T) {
		t.Log("2.2.2.1: onFulfilled must be called after `promise` is fulfilled, with `promise`'s fulfillment value as its first argument.")
		testFulfilled(t, dummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				if value != dummy {
					t.Errorf("got %v, want %v", value, dummy)
				}
				wg.Done()
				return nil
			})
		})
	})
	t.Run("called after promise fulfilled", func(t *testing.T) {
		t.Log("2.2.2.2: onFulfilled must not be called before `promise` is fulfilled.")
		t.Run("fulfill delayed", func(t *testing.T) {
			done := make(chan struct{})
			var isFulfilled atomic.Bool

			p := newPromise()
			p.Then(func(value any) any {
				if !isFulfilled.Load() {
					t.Error("onFulfilled called before promise fulfilled")
				}
				close(done)
				return nil
			})

			go func() {
				time.Sleep(time.Millisecond)
				p.resolve(dummy)
				isFulfilled.Store(true)
			}()

			<-done
		})
		t.Run("never fulfilled", func(t *testing.T) {
			done := make(chan struct{})

			p := newPromise()
			p.Then(func(value any) any {
				close(done)
				return nil
			})

			select {
			case <-done:
				t.Error("got onFulfilled call, want none")
			case <-time.After(time.Millisecond):
				// ok, no call
			}
		})
	})
	t.Run("called once", func(t *testing.T) {
		t.Log("2.2.2.3: onFulfilled must not be called more than once.")
		t.Run("already fulfilled", func(t *testing.T) {
			done := make(chan struct{})
			var timesCalled atomic.Int32

			p := Resolve(dummy)
			p.Then(func(value any) any {
				timesCalled.Add(1)
				if timesCalled.Load() > 1 {
					t.Error("onFulfilled called more than once")
				}
				close(done)
				return nil
			})

			p.resolve(dummy)
			p.resolve(dummy)

			<-done
		})
		t.Run("fulfill immediately", func(t *testing.T) {
			done := make(chan struct{})
			var timesCalled atomic.Int32

			p := newPromise()
			p.Then(func(value any) any {
				timesCalled.Add(1)
				if timesCalled.Load() > 1 {
					t.Error("onFulfilled called more than once")
				}
				close(done)
				return nil
			})

			p.resolve(dummy)
			p.resolve(dummy)

			<-done
		})
		t.Run("fulfill delayed", func(t *testing.T) {
			done := make(chan struct{})
			var timesCalled atomic.Int32

			p := newPromise()
			p.Then(func(value any) any {
				timesCalled.Add(1)
				if timesCalled.Load() > 1 {
					t.Error("onFulfilled called more than once")
				}
				close(done)
				return nil
			})

			go func() {
				time.Sleep(time.Millisecond)
				p.resolve(dummy)
				p.resolve(dummy)
			}()

			<-done
		})
		t.Run("fulfill immediately then delayed", func(t *testing.T) {
			done := make(chan struct{})
			var timesCalled atomic.Int32

			p := newPromise()
			p.Then(func(value any) any {
				timesCalled.Add(1)
				if timesCalled.Load() > 1 {
					t.Error("onFulfilled called more than once")
				}
				close(done)
				return nil
			})

			p.resolve(dummy)
			go func() {
				time.Sleep(time.Millisecond)
				p.resolve(dummy)
			}()

			<-done
		})
		t.Run("multiple then spaced in time", func(t *testing.T) {
			var wg sync.WaitGroup
			var timesCalled [3]atomic.Int32

			wg.Add(4)
			p := newPromise()

			p.Then(func(value any) any {
				timesCalled[0].Add(1)
				if timesCalled[0].Load() > 1 {
					t.Error("1st then: onFulfilled called more than once")
				}
				wg.Done()
				return nil
			})

			go func() {
				time.Sleep(5 * time.Millisecond)
				p.Then(func(value any) any {
					timesCalled[1].Add(1)
					if timesCalled[1].Load() > 1 {
						t.Error("2nd then: onFulfilled called more than once")
					}
					wg.Done()
					return nil
				})
			}()

			go func() {
				time.Sleep(10 * time.Millisecond)
				p.Then(func(value any) any {
					timesCalled[2].Add(1)
					if timesCalled[2].Load() > 1 {
						t.Error("3rd then: onFulfilled called more than once")
					}
					wg.Done()
					return nil
				})
			}()

			go func() {
				time.Sleep(15 * time.Millisecond)
				p.resolve(dummy)
				wg.Done()
			}()

			wg.Wait()
		})
		t.Run("then interleave with fulfillment", func(t *testing.T) {
			var wg sync.WaitGroup
			var timesCalled [2]atomic.Int32

			p := newPromise()
			wg.Add(2)

			p.Then(func(value any) any {
				timesCalled[0].Add(1)
				if timesCalled[0].Load() > 1 {
					t.Error("1st then: onFulfilled called more than once")
				}
				wg.Done()
				return nil
			})

			p.resolve(dummy)

			p.Then(func(value any) any {
				timesCalled[1].Add(1)
				if timesCalled[1].Load() > 1 {
					t.Error("2nd then: onFulfilled called more than once")
				}
				wg.Done()
				return nil
			})

			wg.Wait()
		})
	})
}
