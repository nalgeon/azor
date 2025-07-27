package promise

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestOnRejected(t *testing.T) {
	t.Run("rejection reason", func(t *testing.T) {
		t.Log("2.2.3.1: onRejected must be called after `promise` is rejected, with `promise`'s rejection reason as its first argument.")
		testRejected(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(nil, func(err error) any {
				if !errors.Is(err, errDummy) {
					t.Errorf("got %v, want %v", err, errDummy)
				}
				wg.Done()
				return nil
			})
		})
	})
	t.Run("called after promise rejected", func(t *testing.T) {
		t.Log("2.2.3.2: onRejected must not be called before `promise` is rejected.")
		t.Run("reject delayed", func(t *testing.T) {
			done := make(chan struct{})
			var isRejected atomic.Bool

			p := newPromise()
			p.Then(nil, func(err error) any {
				if !isRejected.Load() {
					t.Error("onRejected called before promise rejected")
				}
				close(done)
				return nil
			})

			go func() {
				time.Sleep(time.Millisecond)
				p.reject(errDummy)
				isRejected.Store(true)
			}()

			<-done
		})
		t.Run("never rejected", func(t *testing.T) {
			done := make(chan struct{})

			p := newPromise()
			p.Then(nil, func(err error) any {
				close(done)
				return nil
			})

			select {
			case <-done:
				t.Error("got onRejected call, want none")
			case <-time.After(time.Millisecond):
				// ok, no call
			}
		})
	})
	t.Run("called once", func(t *testing.T) {
		t.Log("2.2.2.3: onRejected must not be called more than once.")
		t.Run("already rejected", func(t *testing.T) {
			done := make(chan struct{})
			var timesCalled atomic.Int32

			p := Reject(errDummy)
			p.Then(nil, func(err error) any {
				timesCalled.Add(1)
				if timesCalled.Load() > 1 {
					t.Error("onRejected called more than once")
				}
				close(done)
				return nil
			})

			p.reject(errDummy)
			p.reject(errDummy)

			<-done
		})
		t.Run("reject immediately", func(t *testing.T) {
			done := make(chan struct{})
			var timesCalled atomic.Int32

			p := newPromise()
			p.Then(nil, func(err error) any {
				timesCalled.Add(1)
				if timesCalled.Load() > 1 {
					t.Error("onRejected called more than once")
				}
				close(done)
				return nil
			})

			p.reject(errDummy)
			p.reject(errDummy)

			<-done
		})
		t.Run("reject delayed", func(t *testing.T) {
			done := make(chan struct{})
			var timesCalled atomic.Int32

			p := newPromise()
			p.Then(nil, func(err error) any {
				timesCalled.Add(1)
				if timesCalled.Load() > 1 {
					t.Error("onRejected called more than once")
				}
				close(done)
				return nil
			})

			go func() {
				time.Sleep(time.Millisecond)
				p.reject(errDummy)
				p.reject(errDummy)
			}()

			<-done
		})
		t.Run("reject immediately then delayed", func(t *testing.T) {
			done := make(chan struct{})
			var timesCalled atomic.Int32

			p := newPromise()
			p.Then(nil, func(err error) any {
				timesCalled.Add(1)
				if timesCalled.Load() > 1 {
					t.Error("onRejected called more than once")
				}
				close(done)
				return nil
			})

			p.reject(errDummy)
			go func() {
				time.Sleep(time.Millisecond)
				p.reject(errDummy)
			}()

			<-done
		})
		t.Run("multiple then spaced in time", func(t *testing.T) {
			var wg sync.WaitGroup
			var timesCalled [3]atomic.Int32

			p := newPromise()
			wg.Add(4)

			p.Then(nil, func(err error) any {
				timesCalled[0].Add(1)
				if timesCalled[0].Load() > 1 {
					t.Error("1st then: onRejected called more than once")
				}
				wg.Done()
				return nil
			})

			go func() {
				time.Sleep(5 * time.Millisecond)
				p.Then(nil, func(err error) any {
					timesCalled[1].Add(1)
					if timesCalled[1].Load() > 1 {
						t.Error("2nd then: onRejected called more than once")
					}
					wg.Done()
					return nil
				})
			}()

			go func() {
				time.Sleep(10 * time.Millisecond)
				p.Then(nil, func(err error) any {
					timesCalled[2].Add(1)
					if timesCalled[2].Load() > 1 {
						t.Error("3nd then: onRejected called more than once")
					}
					wg.Done()
					return nil
				})
			}()

			go func() {
				time.Sleep(15 * time.Millisecond)
				p.reject(errDummy)
				wg.Done()
			}()

			wg.Wait()
		})
		t.Run("then interleaved with rejection", func(t *testing.T) {
			var wg sync.WaitGroup
			var timesCalled [2]atomic.Int32

			p := newPromise()
			wg.Add(2)

			p.Then(nil, func(err error) any {
				timesCalled[0].Add(1)
				if timesCalled[0].Load() > 1 {
					t.Error("1st then: onRejected called more than once")
				}
				wg.Done()
				return nil
			})

			p.reject(errDummy)

			p.Then(nil, func(err error) any {
				timesCalled[1].Add(1)
				if timesCalled[1].Load() > 1 {
					t.Error("2nd then: onRejected called more than once")
				}
				wg.Done()
				return nil
			})

			wg.Wait()
		})
	})
}
