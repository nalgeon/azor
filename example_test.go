package azor_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nalgeon/azor"
)

func ExampleAsync() {
	calc := func() int { return 42 }

	fn := azor.Async(func() (int, error) {
		return calc(), nil
	})

	ctx := context.Background()
	n, err := azor.Await(ctx, fn())
	fmt.Println(n, err)

	// Output:
	// 42 <nil>
}

func ExamplePromise_Get() {
	p := azor.Run(func() (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 42, nil
	})

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		fmt.Println(p.Get(context.Background()))
		wg.Done()
	}()

	go func() {
		fmt.Println(p.Get(context.Background()))
		wg.Done()
	}()

	wg.Wait()

	// Output:
	// 42 <nil>
	// 42 <nil>
}

func ExamplePromise_Get_cancel() {
	p := azor.Run(func() (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 42, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	val, err := p.Get(ctx)
	fmt.Printf("val = %v, err = %v\n", val, err)

	// Output:
	// val = 0, err = context deadline exceeded
}

func ExampleRun() {
	// Run calls the given function asynchronously
	// and returns a promise.
	p := azor.Run(func() (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 42, nil
	})

	// Get waits for the promise to settle and returns the result.
	val, err := p.Get(context.Background())
	fmt.Printf("val = %v, err = %v\n", val, err)

	// Output:
	// val = 42, err = <nil>
}
