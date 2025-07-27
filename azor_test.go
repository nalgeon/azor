package azor

import (
	"context"
	"errors"
	"testing"
)

type dummy struct {
	value string
}

var valDummy = dummy{value: "dummy"}
var errDummy = errors.New("dummy")

func fnSuccess() (dummy, error) {
	return valDummy, nil
}

func fnError() (int, error) {
	return 0, errDummy
}

func fnPanic() (int, error) {
	panic("oops")
}

func TestAsync(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fn := Async(fnSuccess)
		p := fn()

		val, err := p.Get(t.Context())
		if err != nil {
			t.Errorf("got err = %v, want nil", err)
		}
		if val != valDummy {
			t.Errorf("got val = %v, want %v", val, valDummy)
		}
	})
	t.Run("error", func(t *testing.T) {
		fn := Async(fnError)
		p := fn()

		val, err := p.Get(t.Context())
		if !errors.Is(err, errDummy) {
			t.Errorf("got err = %v, want %v", err, errDummy)
		}
		if val != 0 {
			t.Errorf("got val = %d, want 0", val)
		}
	})
	t.Run("panic", func(t *testing.T) {
		fn := Async(fnPanic)
		p := fn()

		val, err := p.Get(t.Context())
		want := "panic: oops"
		if err == nil || err.Error() != want {
			t.Errorf("got err = %q, want %q", err, want)
		}
		if val != 0 {
			t.Errorf("got val = %d, want 0", val)
		}
	})
	t.Run("nil function", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("should panic for nil function")
			}
		}()
		_ = Async[any](nil)
	})
}

func TestAwait(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := Run(fnSuccess)

		val, err := Await(t.Context(), p)
		if err != nil {
			t.Errorf("got err = %v, want nil", err)
		}
		if val != valDummy {
			t.Errorf("got val = %v, want %v", val, valDummy)
		}
	})
	t.Run("error", func(t *testing.T) {
		p := Run(fnError)

		val, err := Await(t.Context(), p)
		if !errors.Is(err, errDummy) {
			t.Errorf("got err = %v, want %v", err, errDummy)
		}
		if val != 0 {
			t.Errorf("got val = %d, want 0", val)
		}
	})
	t.Run("panic", func(t *testing.T) {
		p := Run(fnPanic)

		val, err := Await(t.Context(), p)
		want := "panic: oops"
		if err == nil || err.Error() != want {
			t.Errorf("got err = %q, want %q", err, want)
		}
		if val != 0 {
			t.Errorf("got val = %d, want 0", val)
		}
	})
	t.Run("canceled", func(t *testing.T) {
		p := Run(fnSuccess)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		val, err := Await(ctx, p)
		if err != ctx.Err() {
			t.Errorf("got err = %v, want %v", err, ctx.Err())
		}
		if val != (dummy{}) {
			t.Errorf("got val = %v, want zero value", val)
		}
	})
	t.Run("nil context", func(t *testing.T) {
		p := Run(fnSuccess)
		val, err := Await(nil, p) // nolint
		if err != nil {
			t.Errorf("got err = %v, want nil", err)
		}
		if val != valDummy {
			t.Errorf("got val = %v, want %v", val, valDummy)
		}
	})
	t.Run("nil promise", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("should panic for nil promise")
			}
		}()
		var p *Promise[any]
		_, _ = Await(t.Context(), p)
	})
}
