package promise

import (
	"sync"
	"testing"
	"time"
)

func TestResolvePromise(t *testing.T) {
	t.Log("2.3.2: If `x` is a promise, adopt its state")
	t.Run("pending", func(t *testing.T) {
		t.Log("2.3.2.1: If `x` is pending, `promise` must remain pending until `x` is fulfilled or rejected.")
		newX := func() *Promise {
			return newPromise()
		}

		testResolution(t, newX, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)

			p.Then(func(value any) any {
				t.Error("onFulfilled should not be called")
				return nil
			}, func(err error) any {
				t.Error("onRejected should not be called")
				return nil
			})

			<-time.After(time.Millisecond)
			wg.Done()
		})
	})
	t.Run("fulfilled", func(t *testing.T) {
		t.Log("2.3.2.2: If/when `x` is fulfilled, fulfill `promise` with the same value.")
		t.Run("already fulfilled", func(t *testing.T) {
			newX := func() *Promise {
				return Resolve(dummy)
			}

			testResolution(t, newX, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(1)

				p.Then(func(value any) any {
					if value != dummy {
						t.Errorf("got %v, want %v", value, dummy)
					}
					wg.Done()
					return nil
				}, func(err error) any {
					t.Error("onRejected should not be called")
					wg.Done()
					return nil
				})
			})
		})
		t.Run("eventually fulfilled", func(t *testing.T) {
			newX := func() *Promise {
				p := newPromise()
				go func() {
					time.Sleep(time.Millisecond)
					p.resolve(dummy)
				}()
				return p
			}

			testResolution(t, newX, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(1)

				p.Then(func(value any) any {
					if value != dummy {
						t.Errorf("got %v, want %v", value, dummy)
					}
					wg.Done()
					return nil
				}, func(err error) any {
					t.Error("onRejected should not be called")
					wg.Done()
					return nil
				})
			})
		})
	})
	t.Run("rejected", func(t *testing.T) {
		t.Log("2.3.2.3: If/when `x` is rejected, reject `promise` with the same reason.")
		t.Run("already rejected", func(t *testing.T) {
			newX := func() *Promise {
				return Reject(errDummy)
			}

			testResolution(t, newX, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(1)

				p.Then(func(value any) any {
					t.Error("onFulfilled should not be called")
					wg.Done()
					return nil
				}, func(err error) any {
					if err != errDummy {
						t.Errorf("got %v, want %v", err, errDummy)
					}
					wg.Done()
					return nil
				})
			})
		})
		t.Run("eventually rejected", func(t *testing.T) {
			newX := func() *Promise {
				p := newPromise()
				go func() {
					time.Sleep(time.Millisecond)
					p.reject(errDummy)
				}()
				return p
			}

			testResolution(t, newX, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(1)

				p.Then(func(value any) any {
					t.Error("onFulfilled should not be called")
					wg.Done()
					return nil
				}, func(err error) any {
					if err != errDummy {
						t.Errorf("got %v, want %v", err, errDummy)
					}
					wg.Done()
					return nil
				})
			})
		})
	})
}
