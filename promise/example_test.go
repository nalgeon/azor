package promise_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nalgeon/azor/promise"
)

type storage map[string]string

func (s storage) Get(ctx context.Context, key string) (string, error) {
	select {
	case <-time.After(10 * time.Millisecond):
		return s[key], nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
func (s storage) Set(ctx context.Context, key, value string) error {
	select {
	case <-time.After(10 * time.Millisecond):
		s[key] = value
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

var db = make(storage)

func ExampleNew() {
	p := promise.New(func(resolve func(any), reject func(error)) {
		// Resolve with a value.
		resolve("go")
	}).Then(func(value any) any {
		// Process the value.
		fmt.Printf("%s is awesome!\n", value)
		return nil
	})
	<-p.Done()

	// Output:
	// go is awesome!
}

func ExamplePromise_cancel() {
	ctx := context.Background()
	p := promise.Resolve("name").Then(func(key any) any {
		ctx, cancel := context.WithTimeout(ctx, time.Millisecond)
		defer cancel()

		val, err := db.Get(ctx, key.(string))
		if err != nil {
			return err
		}
		return val
	}).Then(func(value any) any {
		// Will not be called.
		fmt.Println(value)
		return nil
	}).Catch(func(err error) any {
		// Will be called.
		fmt.Println(err)
		return nil
	})
	<-p.Done()

	// Output:
	// context deadline exceeded
}

func ExamplePromise_chain() {
	p := promise.Resolve(42).Then(func(value any) any {
		// Reject if the value is odd.
		if value.(int)%2 != 0 {
			return fmt.Errorf("odd number: %d", value)
		}
		// Otherwise, print the value.
		fmt.Println("value =", value)
		return value
	}).Catch(func(err error) any {
		// Print the error if it occurs.
		fmt.Println("Error:", err)
		return nil
	}).Finally(func() any {
		// Print a final message.
		fmt.Println("done!")
		return nil
	})
	<-p.Done()

	// Output:
	// value = 42
	// done!
}

func ExamplePromise_Catch() {
	p := promise.New(func(resolve func(any), reject func(error)) {
		// Resolve with a value.
		resolve("go")
	}).Then(func(value any) any {
		// Process the value.
		fmt.Printf("%s is awesome!\n", value)
		return nil
	})
	<-p.Done()

	// Output:
	// go is awesome!
}

func ExamplePromise_Then() {
	p := promise.New(func(resolve func(any), reject func(error)) {
		// Resolve with a value.
		resolve("go")
	}).Then(func(value any) any {
		// Process the value.
		fmt.Printf("%s is awesome!\n", value)
		return nil
	})
	<-p.Done()

	// Output:
	// go is awesome!
}

func ExampleReject() {
	errFailed := fmt.Errorf("failed")
	p := promise.Reject(errFailed).Catch(func(err error) any {
		fmt.Println("Rejected with:", err)
		return nil
	})
	<-p.Done()

	// Output:
	// Rejected with: failed
}

func ExampleResolve() {
	p := promise.Resolve(42).Then(func(value any) any {
		fmt.Println("Resolved with:", value)
		return nil
	})
	<-p.Done()

	// Output:
	// Resolved with: 42
}
