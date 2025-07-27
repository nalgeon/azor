package promise

import (
	"testing"
	"time"
)

func TestRejected(t *testing.T) {
	t.Log("2.1.3.1: When rejected, a promise must not transition to any other state.")
	t.Run("already rejected", func(t *testing.T) {
		onRejectedCalled := make(chan struct{})

		p := Reject(errDummy)
		p.Then(func(value any) any {
			t.Error("got onFullfill call, want none")
			return nil
		}, func(err error) any {
			close(onRejectedCalled)
			return nil
		})

		select {
		case <-onRejectedCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onRejected call, got none")
		}
	})
	t.Run("reject immediately", func(t *testing.T) {
		onRejectedCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			t.Error("got onFulfilled call, want none")
			return nil
		}, func(err error) any {
			close(onRejectedCalled)
			return nil
		})

		p.reject(errDummy)

		select {
		case <-onRejectedCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onRejected call, got none")
		}
	})
	t.Run("reject delayed", func(t *testing.T) {
		onRejectedCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			t.Error("got onFulfilled call, want none")
			return nil
		}, func(err error) any {
			close(onRejectedCalled)
			return nil
		})

		go func() {
			time.Sleep(time.Millisecond)
			p.reject(errDummy)
		}()

		select {
		case <-onRejectedCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onRejected call, got none")
		}
	})
	t.Run("reject then immediately fulfill", func(t *testing.T) {
		onRejectedCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			t.Error("got onFulfilled call, want none")
			return nil
		}, func(err error) any {
			close(onRejectedCalled)
			return nil
		})

		p.reject(errDummy)
		p.resolve(dummy)

		select {
		case <-onRejectedCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onRejected call, got none")
		}
	})
	t.Run("reject then fulfill, delayed", func(t *testing.T) {
		onRejectedCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			t.Error("got onFulfilled call, want none")
			return nil
		}, func(err error) any {
			close(onRejectedCalled)
			return nil
		})

		go func() {
			time.Sleep(time.Millisecond)
			p.reject(errDummy)
			p.resolve(dummy)
		}()

		select {
		case <-onRejectedCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onRejected call, got none")
		}
	})
	t.Run("reject immediately then fulfill delayed", func(t *testing.T) {
		onRejectedCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			t.Error("got onFulfilled call, want none")
			return nil
		}, func(err error) any {
			close(onRejectedCalled)
			return nil
		})

		p.reject(errDummy)
		go func() {
			time.Sleep(time.Millisecond)
			p.resolve(dummy)
		}()

		select {
		case <-onRejectedCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onRejected call, got none")
		}
	})
}
