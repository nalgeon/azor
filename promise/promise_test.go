package promise

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCatch(t *testing.T) {
	t.Run("pending promise", func(t *testing.T) {
		done := make(chan struct{})

		p := newPromise()
		p.Catch(func(err error) any {
			t.Error("onRejected should not be called")
			close(done)
			return nil
		})

		select {
		case <-done:
			t.Error("onRejected should be called")
		case <-time.After(10 * time.Millisecond):
			// ok
		}
	})
	t.Run("fulfilled promise", func(t *testing.T) {
		done := make(chan struct{})

		p := Resolve(dummy)
		p.Catch(func(err error) any {
			t.Error("onRejected should not be called")
			close(done)
			return nil
		})

		select {
		case <-done:
			t.Error("onRejected should be called")
		case <-time.After(10 * time.Millisecond):
			// ok
		}
	})
	t.Run("rejected promise", func(t *testing.T) {
		done := make(chan struct{})

		p := Resolve(errDummy)
		p.Catch(func(err error) any {
			if !errors.Is(err, errDummy) {
				t.Errorf("got %v, want %v", err, errDummy)
			}
			close(done)
			return nil
		})

		select {
		case <-done:
		case <-time.After(10 * time.Millisecond):
			t.Error("onRejected should be called")
		}
	})
}

func TestFinally(t *testing.T) {
	t.Run("fulfilled promise", func(t *testing.T) {
		called := make(chan struct{})

		p := Resolve(dummy)
		p.Finally(func() any {
			close(called)
			return nil
		})

		select {
		case <-called:
		case <-time.After(10 * time.Millisecond):
			t.Error("onFinally should be called")
		}
	})
	t.Run("rejected promise", func(t *testing.T) {
		called := make(chan struct{})

		p := Resolve(errDummy)
		p.Finally(func() any {
			close(called)
			return nil
		})

		select {
		case <-called:
		case <-time.After(10 * time.Millisecond):
			t.Error("onFinally should be called")
		}
	})
	t.Run("chained promise", func(t *testing.T) {
		var called1, called2 atomic.Bool
		done := make(chan struct{})

		p := Resolve(dummy)
		p.Finally(func() any {
			called1.Store(true)
			return nil
		}).Finally(func() any {
			called2.Store(true)
			close(done)
			return nil
		})

		<-done
		if !called1.Load() || !called2.Load() {
			t.Error("onFinally should be called")
		}
	})
	t.Run("panic", func(t *testing.T) {
		done := make(chan struct{})

		p := Resolve(dummy)
		p.Finally(func() any {
			panic("oops")
		}).Then(nil, func(err error) any {
			want := "panic: oops"
			if err.Error() != want {
				t.Errorf("got %q, want %q", err, want)
			}
			close(done)
			return nil
		})

		<-done
	})
	t.Run("return error", func(t *testing.T) {
		done := make(chan struct{})

		p := Resolve(dummy)
		p.Finally(func() any {
			return errDummy
		}).Then(func(value any) any {
			t.Error("should not be fulfilled")
			close(done)
			return nil
		}, func(err error) any {
			if !errors.Is(err, errDummy) {
				t.Errorf("got %v, want %v", err, errDummy)
			}
			close(done)
			return nil
		})

		<-done
	})
	t.Run("return rejected promise", func(t *testing.T) {
		done := make(chan struct{})

		p := Resolve(dummy)
		p.Finally(func() any {
			return Reject(errDummy)
		}).Then(func(value any) any {
			t.Error("should not be fulfilled")
			close(done)
			return nil
		}, func(err error) any {
			if !errors.Is(err, errDummy) {
				t.Errorf("got %v, want %v", err, errDummy)
			}
			close(done)
			return nil
		})

		<-done
	})
	t.Run("return fulfilled promise", func(t *testing.T) {
		done := make(chan struct{})

		p := Resolve(dummy)
		p.Finally(func() any {
			return Resolve("foo")
		}).Then(func(value any) any {
			if value != dummy {
				t.Errorf("got %v, want %v", value, dummy)
			}
			close(done)
			return nil
		}, func(err error) any {
			t.Error("should not be rejected")
			close(done)
			return nil
		})

		<-done
	})
	t.Run("return non-error value", func(t *testing.T) {
		done := make(chan struct{})
		p := Resolve(dummy)
		p.Finally(func() any {
			return "bar"
		}).Then(func(value any) any {
			if value != dummy {
				t.Errorf("got %v, want %v", value, dummy)
			}
			close(done)
			return nil
		}, func(err error) any {
			t.Error("should not be rejected")
			close(done)
			return nil
		})

		<-done
	})
	t.Run("nil handler", func(t *testing.T) {
		done := make(chan struct{})

		p := Resolve(dummy)
		p.Finally(nil).Then(func(value any) any {
			if value != dummy {
				t.Errorf("got %v, want %v", value, dummy)
			}
			close(done)
			return nil
		})

		<-done
	})
}

func TestState(t *testing.T) {
	t.Run("pending", func(t *testing.T) {
		done := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			t.Error("onFulfilled should not be called")
			close(done)
			return nil
		}, func(err error) any {
			t.Error("onRejected should not be called")
			close(done)
			return nil
		})

		select {
		case <-done:
		case <-time.After(10 * time.Millisecond):
			// ok
		}

		select {
		case <-p.Done():
			t.Error("promise should not be settled")
		default:
			// ok
		}
	})
	t.Run("resolved", func(t *testing.T) {
		p := New(func(resolve func(any), reject func(error)) {
			resolve(dummy)
		})
		if p == nil {
			t.Fatal("got nil promise")
		}

		<-p.Done()
		if p.res.err != nil {
			t.Errorf("got err %v, want nil", p.res.err)
		}
		if p.res.val != dummy {
			t.Errorf("got value %v, want %v", p.res.val, dummy)
		}
	})
	t.Run("rejected", func(t *testing.T) {
		p := New(func(resolve func(any), reject func(error)) {
			reject(errDummy)
		})

		<-p.Done()
		if !errors.Is(p.res.err, errDummy) {
			t.Errorf("got err %v, want %v", p.res.err, errDummy)
		}
		if p.res.val != nil {
			t.Errorf("got value %v, want nil", p.res.val)
		}
	})
	t.Run("nil function", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("should panic on nil function")
			}
		}()
		_ = New(nil)
	})
}

func TestPanic(t *testing.T) {
	t.Run("with value", func(t *testing.T) {
		done := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			panic("oops")
		}).Then(nil, func(err error) any {
			want := "panic: oops"
			if err.Error() != want {
				t.Errorf("got %q, want %q", err, want)
			}
			close(done)
			return nil
		})
		p.resolve(dummy)

		<-done
	})
	t.Run("with error", func(t *testing.T) {
		done := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			panic(errDummy)
		}).Then(nil, func(err error) any {
			if !errors.Is(err, errDummy) {
				t.Errorf("got %v, want %v", err, errDummy)
			}
			close(done)
			return nil
		})
		p.resolve(dummy)

		<-done
	})
	t.Run("on fulfilled", func(t *testing.T) {
		testFulfilled(t, dummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)

			p.Then(func(value any) any {
				panic("oops")
			}).Then(func(value any) any {
				t.Error("onFulfilled should not be called")
				wg.Done()
				return nil
			}, func(err error) any {
				want := "panic: oops"
				if err.Error() != want {
					t.Errorf("got %q, want %q", err, want)
				}
				wg.Done()
				return nil
			})
		})
	})
	t.Run("on rejected", func(t *testing.T) {
		testRejected(t, errDummy, func(t *testing.T, p *Promise, wg *sync.WaitGroup) {
			wg.Add(1)

			p.Then(nil, func(err error) any {
				panic("oops")
			}).Then(func(value any) any {
				t.Error("onFulfilled should not be called")
				wg.Done()
				return nil
			}, func(err error) any {
				want := "panic: oops"
				if err.Error() != want {
					t.Errorf("got %q, want %q", err, want)
				}
				wg.Done()
				return nil
			})
		})
	})
}

func TestReject(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		p := Reject(errDummy)
		select {
		case <-p.done:
			// ok
		default:
			t.Error("want a settled promise")
		}
		if p.res.err == nil {
			t.Errorf("got nil err, want %v", errDummy)
		}
		if p.res.val != nil {
			t.Errorf("got value %v, want nil", p.res.val)
		}
	})
	t.Run("nil", func(t *testing.T) {
		p := Reject(nil)
		select {
		case <-p.done:
			// ok
		default:
			t.Error("want a settled promise")
		}
		if p.res.err != nil {
			t.Errorf("got err %v, want nil", p.res.err)
		}
		if p.res.val != nil {
			t.Errorf("got value %v, want nil", p.res.val)
		}
	})
}

func TestResolve(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		p := Resolve(dummy)
		select {
		case <-p.done:
			// ok
		default:
			t.Error("want a settled promise")
		}
		if p.res.err != nil {
			t.Errorf("got err %v, want nil", p.res.err)
		}
		if p.res.val != dummy {
			t.Errorf("got value %v, want %v", p.res.val, dummy)
		}
	})
	t.Run("error", func(t *testing.T) {
		p := Resolve(errDummy)
		select {
		case <-p.done:
			// ok
		default:
			t.Error("want a settled promise")
		}
		if !errors.Is(p.res.err, errDummy) {
			t.Errorf("got err %v, want %v", p.res.err, errDummy)
		}
		if p.res.val != nil {
			t.Errorf("got value %v, want nil", p.res.val)
		}
	})
	t.Run("nil", func(t *testing.T) {
		p := Resolve(nil)
		select {
		case <-p.done:
			// ok
		default:
			t.Error("want a settled promise")
		}
		if p.res.err != nil {
			t.Errorf("got err %v, want nil", p.res.err)
		}
		if p.res.val != nil {
			t.Errorf("got value %v, want nil", p.res.val)
		}
	})
	t.Run("promise", func(t *testing.T) {
		p1 := Resolve(dummy)
		p2 := Resolve(p1)
		select {
		case <-p2.done:
			// ok
		default:
			t.Error("want a settled promise")
		}
		if p2.res.err != nil {
			t.Errorf("got err %v, want nil", p2.res.err)
		}
		if p2.res.val != dummy {
			t.Errorf("got value %v, want %v", p2.res.val, dummy)
		}
	})
}
