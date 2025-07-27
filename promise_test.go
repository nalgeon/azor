package azor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestPromiseGet(t *testing.T) {
	t.Run("block until completion", func(t *testing.T) {
		started := make(chan struct{})
		done := make(chan struct{})

		p := Run(func() (int, error) {
			<-started
			return 42, nil
		})

		start := time.Now()
		go func() {
			val, err := p.Get(context.Background())
			if time.Since(start) < 10*time.Millisecond {
				t.Error("should block until the future is completed")
			}
			if err != nil {
				t.Errorf("got err = %v, want nil", err)
			}
			if val != 42 {
				t.Errorf("got val = %d, want 42", val)
			}
			close(done)
		}()

		time.Sleep(10 * time.Millisecond)
		close(started)
		<-done
	})
	t.Run("nil context", func(t *testing.T) {
		p := Run(func() (int, error) {
			return 42, nil
		})
		val, err := p.Get(nil) // nolint
		if err != nil {
			t.Errorf("got err = %v, want nil", err)
		}
		if val != 42 {
			t.Errorf("got val = %d, want 42", val)
		}
	})
	t.Run("timeout", func(t *testing.T) {
		started := make(chan struct{})
		p := Run(func() (int, error) {
			<-started
			return 42, nil
		})

		go func() {
			time.Sleep(10 * time.Millisecond)
			<-started
		}()

		ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond)
		defer cancel()
		val, err := p.Get(ctx)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("got err = %v, want context.DeadlineExceeded", err)
		}
		if val != 0 {
			t.Errorf("got val = %d, want 0", val)
		}
	})
	t.Run("concurrent get", func(t *testing.T) {
		p := Run(func() (int, error) {
			return 42, nil
		})

		var vals [2]int
		var errs [2]error

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			vals[0], errs[0] = p.Get(t.Context())
			wg.Done()
		}()

		go func() {
			vals[1], errs[1] = p.Get(t.Context())
			wg.Done()
		}()

		wg.Wait()
		if errs[0] != nil || errs[1] != nil {
			t.Errorf("got errs = %v, want nil", errs)
		}
		if vals[0] != 42 || vals[1] != 42 {
			t.Errorf("got vals = %v, want [42 42]", vals)
		}
	})
}

func TestPromiseState(t *testing.T) {
	t.Run("pending", func(t *testing.T) {
		start := make(chan struct{})
		p := Run(func() (int, error) {
			<-start
			return 42, nil
		})

		if p == nil {
			t.Fatal("got nil future")
		}
		select {
		case <-p.Done():
			t.Error("done should not be closed")
		default:
			//  ok
		}

		close(start)
	})
	t.Run("completed", func(t *testing.T) {
		p := Run(func() (int, error) {
			return 42, nil
		})

		_, _ = p.Get(t.Context())
		select {
		case <-p.Done():
			//  ok
		default:
			t.Error("done should be closed")
		}
	})
}

func TestRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := Run(func() (int, error) {
			return 42, nil
		})

		val, err := p.Get(t.Context())
		if err != nil {
			t.Errorf("got err = %v, want nil", err)
		}
		if val != 42 {
			t.Errorf("got val = %d, want 42", val)
		}
	})
	t.Run("error", func(t *testing.T) {
		var errDummy = errors.New("dummy")
		p := Run(func() (int, error) {
			return 0, errDummy
		})

		val, err := p.Get(t.Context())
		if !errors.Is(err, errDummy) {
			t.Errorf("got err = %v, want %v", err, errDummy)
		}
		if val != 0 {
			t.Errorf("got val = %d, want 0", val)
		}
	})
	t.Run("panic value", func(t *testing.T) {
		p := Run(func() (int, error) {
			panic("oops")
		})

		val, err := p.Get(t.Context())
		want := "panic: oops"
		if err == nil || err.Error() != want {
			t.Errorf("got err = %q, want %q", err, want)
		}
		if val != 0 {
			t.Errorf("got val = %d, want 0", val)
		}
	})
	t.Run("panic error", func(t *testing.T) {
		errDummy := errors.New("dummy")
		p := Run(func() (int, error) {
			panic(errDummy)
		})

		val, err := p.Get(t.Context())
		if !errors.Is(err, errDummy) {
			t.Errorf("got err = %q, want %q", err, errDummy)
		}
		if val != 0 {
			t.Errorf("got val = %d, want 0", val)
		}
	})
	t.Run("canceled", func(t *testing.T) {
		started := make(chan struct{})
		ctx, cancel := context.WithCancel(t.Context())
		p := Run(func() (int, error) {
			<-started
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
				return 42, nil
			}
		})

		cancel()
		close(started)

		val, err := p.Get(t.Context())
		if !errors.Is(err, ctx.Err()) {
			t.Errorf("got err = %v, want %v", err, ctx.Err())
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
		Run[int](nil)
	})
}
