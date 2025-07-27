package promise

import (
	"testing"
	"time"
)

func TestFulfilled(t *testing.T) {
	t.Log("2.1.2.1: When fulfilled, a promise must not transition to any other state.")
	t.Run("already fulfilled", func(t *testing.T) {
		onFulfilledCalled := make(chan struct{})

		p := Resolve(dummy)
		p.Then(func(value any) any {
			close(onFulfilledCalled)
			return nil
		}, func(err error) any {
			t.Error("got onRejected call, want none")
			return nil
		})

		select {
		case <-onFulfilledCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onFulfilled call, got none")
		}
	})
	t.Run("fulfill immediately", func(t *testing.T) {
		onFulfilledCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			close(onFulfilledCalled)
			return nil
		}, func(err error) any {
			t.Error("got onRejected call, want none")
			return nil
		})

		p.resolve(dummy)

		select {
		case <-onFulfilledCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onFulfilled call, got none")
		}
	})
	t.Run("fulfill delayed", func(t *testing.T) {
		onFulfilledCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			close(onFulfilledCalled)
			return nil
		}, func(err error) any {
			t.Error("got onRejected call, want none")
			return nil
		})

		go func() {
			time.Sleep(time.Millisecond)
			p.resolve(dummy)
		}()

		select {
		case <-onFulfilledCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onFulfilled call, got none")
		}
	})
	t.Run("fulfill then immediately reject", func(t *testing.T) {
		onFulfilledCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			close(onFulfilledCalled)
			return nil
		}, func(err error) any {
			t.Error("got onRejected call, want none")
			return nil
		})

		p.resolve(dummy)
		p.reject(errDummy)

		select {
		case <-onFulfilledCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onFulfilled call, got none")
		}
	})
	t.Run("fulfill then reject, delayed", func(t *testing.T) {
		onFulfilledCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			close(onFulfilledCalled)
			return nil
		}, func(err error) any {
			t.Error("got onRejected call, want none")
			return nil
		})

		go func() {
			time.Sleep(time.Millisecond)
			p.resolve(dummy)
			p.reject(errDummy)
		}()

		select {
		case <-onFulfilledCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onFulfilled call, got none")
		}
	})
	t.Run("fulfill immediately then reject delayed", func(t *testing.T) {
		onFulfilledCalled := make(chan struct{})

		p := newPromise()
		p.Then(func(value any) any {
			close(onFulfilledCalled)
			return nil
		}, func(err error) any {
			t.Error("got onRejected call, want none")
			return nil
		})

		p.resolve(dummy)
		go func() {
			time.Sleep(time.Millisecond)
			p.reject(errDummy)
		}()

		select {
		case <-onFulfilledCalled:
			// ok
		case <-time.After(10 * time.Millisecond):
			t.Error("want onFulfilled call, got none")
		}
	})
}
