package promise

import (
	"errors"
	"sync"
	"testing"
)

func TestMultiThen(t *testing.T) {
	t.Log("2.2.6: `then` may be called multiple times on the same promise")
	t.Run("on fulfilled", func(t *testing.T) {
		t.Run("multiple handlers", func(t *testing.T) {
			testFulfilled(t, dummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(3)
				onFulfilled1 := func(value any) any {
					if value != dummy {
						t.Errorf("onFulfilled1: got %v, want %v", value, dummy)
					}
					wg.Done()
					return nil
				}
				onFulfilled2 := func(value any) any {
					if value != dummy {
						t.Errorf("onFulfilled2: got %v, want %v", value, dummy)
					}
					wg.Done()
					return nil
				}
				onFulfilled3 := func(value any) any {
					if value != dummy {
						t.Errorf("onFulfilled3: got %v, want %v", value, dummy)
					}
					wg.Done()
					return nil
				}
				onRejected := func(err error) any {
					t.Error("onRejected should not be called")
					return nil
				}

				p.Then(onFulfilled1, onRejected)
				p.Then(onFulfilled2, onRejected)
				p.Then(onFulfilled3, onRejected)
			})
		})
		t.Run("handlers with errors", func(t *testing.T) {
			testFulfilled(t, dummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(3)
				onFulfilled1 := func(value any) any {
					if value != dummy {
						t.Errorf("onFulfilled1: got %v, want %v", value, dummy)
					}
					wg.Done()
					return nil
				}
				onFulfilled2 := func(value any) any {
					if value != dummy {
						t.Errorf("onFulfilled2: got %v, want %v", value, dummy)
					}
					wg.Done()
					return errors.New("onFulfilled2")
				}
				onFulfilled3 := func(value any) any {
					if value != dummy {
						t.Errorf("onFulfilled3: got %v, want %v", value, dummy)
					}
					wg.Done()
					return nil
				}
				onRejected := func(err error) any {
					t.Error("onRejected should not be called")
					return nil
				}

				p.Then(onFulfilled1, onRejected)
				p.Then(onFulfilled2, onRejected)
				p.Then(onFulfilled3, onRejected)
			})
		})
		t.Run("multiple chains", func(t *testing.T) {
			testFulfilled(t, dummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(3)

				sentinel1 := struct{ sentinel1 string }{"sentinel1"}
				p.Then(func(value any) any {
					return sentinel1
				}).Then(func(value any) any {
					if value != sentinel1 {
						t.Errorf("chain 1: got %v, want %v", value, sentinel1)
					}
					wg.Done()
					return nil
				})

				err2 := errors.New("err2")
				p.Then(func(value any) any {
					return err2
				}).Then(nil, func(err error) any {
					if !errors.Is(err, err2) {
						t.Errorf("chain 2: got %v, want %v", err, err2)
					}
					wg.Done()
					return nil
				})

				sentinel3 := struct{ sentinel3 string }{"sentinel3"}
				p.Then(func(value any) any {
					return sentinel3
				}).Then(func(value any) any {
					if value != sentinel3 {
						t.Errorf("chain1: got %v, want %v", value, sentinel3)
					}
					wg.Done()
					return nil
				})
			})
		})
		// NOT IMPLEMENTED:
		// 2.2.6.1: If/when promise is fulfilled, all respective onFulfilled callbacks
		// must execute in the order of their originating calls to then.
	})
	t.Run("on rejected", func(t *testing.T) {
		t.Run("multiple handlers", func(t *testing.T) {
			testRejected(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(3)

				onRejected1 := func(err error) any {
					if !errors.Is(err, errDummy) {
						t.Errorf("onRejected1: got %v, want %v", err, errDummy)
					}
					wg.Done()
					return nil
				}
				onRejected2 := func(err error) any {
					if !errors.Is(err, errDummy) {
						t.Errorf("onRejected1: got %v, want %v", err, errDummy)
					}
					wg.Done()
					return nil
				}
				onRejected3 := func(err error) any {
					if !errors.Is(err, errDummy) {
						t.Errorf("onRejected1: got %v, want %v", err, errDummy)
					}
					wg.Done()
					return nil
				}
				onFulfilled := func(value any) any {
					t.Error("onFulfilled should not be called")
					return nil
				}

				p.Then(onFulfilled, onRejected1)
				p.Then(onFulfilled, onRejected2)
				p.Then(onFulfilled, onRejected3)
			})
		})
		t.Run("handlers with errors", func(t *testing.T) {
			testRejected(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(3)

				onRejected1 := func(err error) any {
					if !errors.Is(err, errDummy) {
						t.Errorf("onRejected1: got %v, want %v", err, errDummy)
					}
					wg.Done()
					return nil
				}
				onRejected2 := func(err error) any {
					if !errors.Is(err, errDummy) {
						t.Errorf("onRejected1: got %v, want %v", err, errDummy)
					}
					wg.Done()
					return errors.New("onRejected2")
				}
				onRejected3 := func(err error) any {
					if !errors.Is(err, errDummy) {
						t.Errorf("onRejected1: got %v, want %v", err, errDummy)
					}
					wg.Done()
					return nil
				}
				onFulfilled := func(value any) any {
					t.Error("onFulfilled should not be called")
					return nil
				}

				p.Then(onFulfilled, onRejected1)
				p.Then(onFulfilled, onRejected2)
				p.Then(onFulfilled, onRejected3)
			})
		})
		t.Run("multiple chains", func(t *testing.T) {
			testRejected(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
				wg.Add(3)

				sentinel1 := struct{ sentinel1 string }{"sentinel1"}
				p.Then(nil, func(err error) any {
					return sentinel1
				}).Then(func(value any) any {
					if value != sentinel1 {
						t.Errorf("chain 1: got %v, want %v", value, sentinel1)
					}
					wg.Done()
					return nil
				})

				err2 := errors.New("err2")
				p.Then(nil, func(err error) any {
					return err2
				}).Then(nil, func(err error) any {
					if !errors.Is(err, err2) {
						t.Errorf("chain 2: got %v, want %v", err, err2)
					}
					wg.Done()
					return nil
				})

				sentinel3 := struct{ sentinel3 string }{"sentinel3"}
				p.Then(nil, func(err error) any {
					return sentinel3
				}).Then(func(value any) any {
					if value != sentinel3 {
						t.Errorf("chain 3: got %v, want %v", value, sentinel3)
					}
					wg.Done()
					return nil
				})
			})
		})
		// NOT IMPLEMENTED:
		// 2.2.6.2: If/when `promise` is rejected, all respective `onRejected` callbacks
		// must execute in the order of their originating calls to `then`.
	})
}
