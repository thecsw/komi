# Komi - subarashi go pooling 🍡

<div id='badges' align='center'>

[![Go Report Card](https://goreportcard.com/badge/github.com/thecsw/komi)](https://goreportcard.com/report/github.com/thecsw/komi)
[![GoDoc](https://godoc.org/github.com/thecsw/komi?status.svg)](https://godoc.org/github.com/thecsw/komi)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

</div>

![komi](komi.jpg)

## Motivation

Go is great for setting up easy parallel jobs and processes, however, it is not easy
when one starts confusing concurrency with parallelism and ending up endlessly fighting
race conditions. `komi` is a generic pooling library that will satisfy your hunger.

## Usage

Say you want to run a function `foo(v)` that performs some kind of work on parameter `v`,
be it a database operation, a syscall, an IO operation, etc. (possibilities are endless!)
Setting up a pool and sending jobs is as trivial as

```go
pool := komi.New(komi.WorkSimple(foo))
defer pool.Close()
// other code...
err = pool.Submit(v) // will block if pool is full
```

Notice that `pool.Close()` will gracefully free all the resources and channels occupied by
`pool` by waiting for final jobs to complete. `pool.Close(true)` will force pool closure.

But what if you want to collect outputs of work performed on `v` with `foo(v) w`?

```go
pool := komi.New(komi.Work(foo))
defer pool.Close()
// collect outputs with pool.Outputs() channel
go func() {
	// Errors omitted for brevity
	outputs, _ := pool.Outputs()
	for output := range outputs {
		// output is the result of `foo(v)`
	}
}()
// other code...
err = pool.Submit(v) // will block if pool is full
```

But what if you want to collect errors as well? Consider `foo(v) error `

```go
pool := komi.New(komi.WorkSimpleWithErrors(foo))
defer pool.Close()
// collect errors with with pool.Errors() channel...
// other code...
err = pool.Submit(v) // will block if pool is full
```

Or with `foo(v) (w, error)`!

```go
pool := komi.New(komi.WorkWithErrors(foo))
defer pool.Close()
// collect outputs with pool.Outputs() channel
// collect errors with pool.Errors() channel...
// other code...
err = pool.Submit(v) // will block if pool is full
```

So, depending on what function you give, any work type is handled by the pool
on the fly! If work given doesn't produce outputs, `pool.Outputs()` will return `nil`,
similarly, if work given doesn't produce errors, `pool.Errors()` will return `nil`.

Note: if work produces outputs or errors, those activated channels **need** to be consumed
by the user, otherwise, when reaching `size` number of elements in either (if active), work
will be blocked until the destination channel is consumed.

## Connectors

Unique feature of `komi` is that each pool can be connected with each other. Say you have two
functions, where one opens file's contents, `openFile(filename string) (string, error)`,
and the other counts the number of words, `countWords(contents string) int`.

Two pools can be created,

```go
opener := komi.NewWithSettings(komi.WorkWithErrors(openFile), &komi.Settings{
	Name:     "Opener 📂 ",
	Laborers: 1,
	Size:     4,
})

counter := komi.NewWithSettings(komi.Work(countWords), &komi.Settings{
	Name:     "Counter 📚 ",
	Laborers: 10,
	Size:     20,
})
```

We can wire the outputs of `opener` to be automatically fed into `counter` with

```go
opener.Connect(counter)
```

So now, those two pools are "connected". We would call this relationship as `opener` being
the dependent (child) pool and `counter` being the connected (parent) pool.

```
              Opener 📂                   Counter 📚
filenames  ┌─────────────┐  contents   ┌──────────────┐  word counts
 ───────>  │  openFile   │  ────────>  │  countWords  │  ────────>
 .Submit   └─────────────┘  .Connect   └──────────────┘  .Outputs
            pool: opener                 pool: counter 
```

`komi`'s pools are smart enough that only by calling `counter.Close()`, it will issue a shutdown
command back to `opener` and wait until it's closed. This closing logic procedure will happen with
any number of connected pools.

If you have pools `1,2,...,N-1,N` connected in form `1->2->...->N-1->N`, user **can only call**
`.Close()` on pool `N`, as it would be responsible for sending a closure request to `N-1`, and 
so on until `2` sends the shutdown request to `1`. When `1` is closed, the closure will resume
on `2`, up until `N-1` and `N`, where the latter will return from the original `Close()` call.

Please note that none of the pools `1,2,...,N-1` in the above will honor user's closure request,
as it should come from their connected (parent) pool.

## Quirks

When the parent-most pool is closing, it will wait for all the child pools to complete their jobs.
This behavior can be overwritten by calling `.Close(true)`, which would skip any waiting of childrens'
queued jobs and skip any waiting of the parent-most pool's queued jobs.

## Operations

Pools support waiting (blocking) until the pool has no jobs waiting for completion with `pool.Wait()`.

Some other quality of life operations are also provided,

- `Submit(v)` will submit job `v` to be performed by the pool. 
- `Close()` will close the pool if and only if it's disconnected or the parent-most pool.
- `Close(true)` will close the pool ignoring any pending jobs.
- `Outputs()` will return channel that the user should listen to for outputs (if work generated them).
- `Errors()` will return channel that the user shoud listen to for errors (if work generates them).
- `IsConnected()` will return true if the pool is a child of another pool, thus sending its outputs.
- `IsClosed()` will return true if the pool has gracefully shutdown.
- `JobsCompleted()` will return the number of jobs this pool has completed.
- `JobsWaiting()` will return the number of jobs waiting in queue and currently in-work.
- `JobsSucceeded()` will return the number of jobs completed with a non-nil errors.
- `Name()` will return the pool's name (defaults to `Komi 🍡 `).

## Settings

You can tune the performance and behavior of the pool with `komi.NewWithSetttings` by providing `*komi.Settings`,

- `Laborers` sets the number of pool's laborers.
- `Size` sets the size of the pool (how many jobs can wait until `pool.Submit` is blocked).
- `Ratio` sets the `ratio` in `size = ratio * number of laborers` equation (only if size has not been manually set).
- `LogLevel` sets the pool's logging level to `level`.
- `Debug` sets the pool's logging level to `DebugLevel`.
- `Name` sets the pool's name as shown in logs.

## Stability

This is a brand new library I built for my [static website generator](https://github.com/thecsw/darkness),
where it's used extensively and in production. However, there are no guarantees provided for this library,
that is, until something like `v1.0` is out, in which case, I would promise to maintain backward compatibility.

Please use it with knowing your risks. However, if you use a tagged version or a commit hash in your `go.mod`,
you should be fine.

## Future work

Some future items in mind:

- Adding an error handler to `komi` polls, which if given, will be invoked if non-nil errors are returned by work.
- More tests

## Developers

Please consider giving it a try and filing an issue or a pull request.

Thank you!
