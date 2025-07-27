# Azor - promises and async/await in Go

Azor offers a promises implementation that mostly follows the Promises/A+ specification, along with async/await features similar to those in JavaScript.

> I know that promises and async/await aren't how you write concurrent programs in Go. This is a research project — I was curious to see if I could create JavaScript-style promises using Go's concurrency primitives. The implementation turned out to be surprisingly simple and clean.

## Promises

`promise.Promise` ([source](https://github.com/nalgeon/azor/blob/main/promise/promise.go#L25)) is a JavaScript-like promise that executes a given function in a separate goroutine:

```go
promise.New(func(resolve func(any), reject func(error)) {
    time.Sleep(10 * time.Millisecond)
    resolve("done")
})
```

You can wait for a promise to complete (settle) with `Done`:

```go
p := promise.New(func(resolve func(any), reject func(error)) {
    time.Sleep(10 * time.Millisecond)
    resolve("done")
})

<-p.Done()
```

The promise is eventually either _fulfilled_ (successful result) or _rejected_ (an error or a panic). Both cases are handled by `Then` callbacks:

```go
promise.New(func(resolve func(any), reject func(error)) {
    // Resolve or reject based on the random value.
    if n := rand.N(100); n%2 == 0 {
        resolve(n)
    } else {
        reject(fmt.Errorf("odd number: %v", n))
    }
}).Then(func(value any) any {
    // The original promise was fulfilled.
    fmt.Println(value)
    return nil
}, func(err error) any {
    // The original promise was rejected.
    fmt.Println(err)
    return nil
})

// Output:
// odd number: 21
```

You can build multi-step flows using `Then`, `Catch` and `Finally`:

```go
promise.Resolve(42).Then(func(value any) any {
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

// Output:
// value = 42
// done!
```

To handle cancellation, use a context as usual and return the context's error:

```go
ctx := context.Background()
promise.Resolve("name").Then(func(key any) any {
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

// Output:
// context deadline exceeded
```

## Asynchronous computing

The top-level package offers a simple, type-safe `Promise[T]` ([source](https://github.com/nalgeon/azor/blob/main/promise.go#L10)) that runs a given function asynchronously and returns the result, without including all the extra features from the official spec:

```go
// Run calls the given function asynchronously and returns a promise.
p := azor.Run(func() (int, error) {
    time.Sleep(10 * time.Millisecond)
    return 42, nil
})

// Get waits for the promise to settle and returns the result.
val, err := p.Get(context.Background())
fmt.Printf("val = %v, err = %v\n", val, err)

// Output:
// val = 42, err = <nil>
```

You can safely call `Get` from multiple goroutines. It will always return the same result:

```go
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
```

Use a context to set the wait timeout:

```go
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
```

## Async/await

With Azor, you get all the (highly questionable) benefits of async/await without the "viral" effects of using the `async` keyword. Write a regular function:

```go
calc := func() int { return 42 }
```

Turn it into an async function:

```go
fn := azor.Async(func() (int, error) {
    return calc(), nil
})
```

Then call it asynchronously and get the result:

```go
ctx := context.Background()
n, err := azor.Await(ctx, fn())
fmt.Println(n, err)

// Output:
// 42 <nil>
```

`Async` and `Await` ([source](https://github.com/nalgeon/azor/blob/main/azor.go#L8)) are just convenience wrappers for `Run` and `Promise`:

-   `Async` returns a function that calls `Run` and gives back a `Promise` when you call it.
-   `Await` calls `Promise.Get` on this promise to retrieve the result.

## Specification compliance

`promise.Promise` follows the [Promises/A+](https://promisesaplus.com) specification, with two exceptions:

➊ If you call `Then` multiple times on the same promise, the handlers may run in any order.

The original specification requires that separate `Then` handlers run sequentially, in the order that `Then` was called:

```go
p := promise.Resolve()
p.Then(fn1)
p.Then(fn2)
p.Then(fn3)
// fn1, fn2 and fn3 are required to run in this exact order.
```

This doesn't make much sense for a truly concurrent runtime like Go. In Azor's implementation, the handlers can run in any order. If you want to control the order, you can still do it by chaining them:

```go
p := promise.Resolve()
p.Then(fn1).Then(fn2).Then(fn3)
// fn1, fn2 and fn3 are guaranteed to run in this exact order.
```

➋ There is no special handling for "thenables" (objects with a "then" method).

Thenables in the original specification were basically a workaround to support all the different promise implementations that existed before the spec was created (like jQuery's promise). Since Go doesn't have legacy promises (or any promises at all), it made no sense to support thenables in Azor.

## Frequently asked questions

> Why?

To see how difficult it is to implement a "foreign" concurrency approach in Go.

> Why not use generics?

In short, Go's generics aren't powerful enough to support the original spec. For example, the `Then` method can change the type of the promise value, which would require its own type parameter for the method. But in Go, methods can only use type parameters defined on the type itself:

```go
type Promise[V any] struct {}

// allowed
func (p *Promise[V]) Then(handler func(V) V) *Promise[V] {}

// not allowed
func (p *Promise[V, NV any]) Then(handler func(V) NV) *Promise[NV] {}
```

Still, there is a type-safe `Promise[T]` wrapper in the top-level package that you can use with `Run` or `Async`/`Await` as described above. It doesn't support chaining (which is a good thing, in my opinion).

> Is it production-ready?

Pretty much. The code is reasonably simple and readable, and the test coverage is good (I've ported the original spec test suite to Go).

> Should I use it?

Absolutely not. Promises are an unnecessarily complicated way to solve problems that are much better handled with Go's standard concurrency tools.

But if you decide to use it, I strongly recommend sticking to the top-level package types and functions like `Promise[T]` and `Run`.

## Contributing

Contributions are welcome. For anything other than bugfixes, please first open an issue to discuss what you want to change.

Make sure to add or update tests as needed.

## License

Created by [Anton Zhiyanov](https://antonz.org/). Released under the MIT License.

Logo by [Lorc](https://lorcblog.blogspot.com/) under [CC BY 3.0](http://creativecommons.org/licenses/by/3.0/).
