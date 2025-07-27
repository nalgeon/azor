package promise

import (
	"errors"
	"sync"
	"testing"
)

func TestResolveValue(t *testing.T) {
	t.Log("2.3.3.4: If `then` is not a function, fulfill `promise` with `x`.")
	t.Log("2.3.4: If `x` is not an object or function, fulfill `promise` with `x`.")
	t.Run("nil", func(t *testing.T) {
		testFulfilled(t, nil, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				if value != nil {
					t.Errorf("got %v, want nil", value)
				}
				wg.Done()
				return nil

			}, func(err error) any {
				t.Error("onRejected should not be called")
				wg.Done()
				return nil
			})
		})
		testRejected(t, nil, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				if value != nil {
					t.Errorf("got %v, want nil", value)
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
	t.Run("bool", func(t *testing.T) {
		testFulfilled(t, true, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				if value != true {
					t.Errorf("got %v, want true", value)
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
	t.Run("int", func(t *testing.T) {
		testFulfilled(t, 42, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				if value != 42 {
					t.Errorf("got %v, want 42", value)
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
	t.Run("struct", func(t *testing.T) {
		testFulfilled(t, dummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
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
	t.Run("pointer", func(t *testing.T) {
		sentinel := &dummy
		testFulfilled(t, sentinel, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				if value != sentinel {
					t.Errorf("got %v, want %v", value, sentinel)
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
	t.Run("interface", func(t *testing.T) {
		sentinel := any(dummy)
		testFulfilled(t, sentinel, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				if value != sentinel {
					t.Errorf("got %v, want %v", value, sentinel)
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
	t.Run("func", func(t *testing.T) {
		fn := func() int { return 42 }
		testFulfilled(t, fn, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				gotFn, ok := value.(func() int)
				if !ok {
					t.Fatal("want func() int")
				}
				if got := gotFn(); got != 42 {
					t.Errorf("fn() = %v, want 42", got)
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
	t.Run("error", func(t *testing.T) {
		testFulfilled(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				t.Error("onFulfilled should not be called")
				wg.Done()
				return nil
			}, func(err error) any {
				if !errors.Is(err, errDummy) {
					t.Errorf("got %v, want %v", err, errDummy)
				}
				wg.Done()
				return nil
			})
		})
		testRejected(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)
			p.Then(func(value any) any {
				t.Error("onFulfilled should not be called")
				wg.Done()
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
}
