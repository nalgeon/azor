package promise

import (
	"errors"
	"testing"
)

func TestResolveSelf(t *testing.T) {
	t.Log("2.3.1: If `promise` and `x` refer to the same object, reject `promise` with a `TypeError' as the reason.")
	t.Run("on fulfilled", func(t *testing.T) {
		done := make(chan struct{})

		p := Resolve(dummy)
		p.Then(func(value any) any {
			return p
		}).Then(func(value any) any {
			t.Error("onFulfilled should not be called")
			close(done)
			return nil
		}, func(err error) any {
			if !errors.Is(err, errors.ErrUnsupported) {
				t.Errorf("got %v, want %v", err, errors.ErrUnsupported)
			}
			close(done)
			return nil
		})

		<-done
	})
	t.Run("on rejected", func(t *testing.T) {
		done := make(chan struct{})

		p := Reject(errDummy)
		p.Then(nil, func(err error) any {
			return p
		}).Then(func(value any) any {
			t.Error("onFulfilled should not be called")
			close(done)
			return nil
		}, func(err error) any {
			if !errors.Is(err, errors.ErrUnsupported) {
				t.Errorf("got %v, want %v", err, errors.ErrUnsupported)
			}
			close(done)
			return nil
		})

		<-done
	})
}
