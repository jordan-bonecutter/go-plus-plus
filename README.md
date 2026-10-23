# Go++

Go but I am the captain now

## What did you add?

I added tagged unions, a la:

```go
type Option[T any] union {
    None
    Some T
}

func valueOr[T any](option Option[T], fallback T) T {
    switch value := option.(variant) {
    case Some:
        return value // T
    case None:
        return fallback
    }
}
```

Variant switch statements must be exhaustive. You can also extract a specific variant a-la:
```go
value, ok := x.(Some)
```

Construct values with `Option.Some(42)` or `Option.None()`.
Omitting a variant's type makes it a unit, which is an alias for `struct{}`.
Payloads overlap inline. Pointer-free unions omit the GC anchor, including
per generic instantiation: `Option[int]` needs none; `Option[string]` does.
Unions have their own `reflect.Union` kind and support JSON, gob, and XML.

## Error propagation

```go
func readConfig(path string) (map[string]any, error) {
    data := os.ReadFile(path)!
    var config map[string]any
    json.Unmarshal(data, &config)!
    return config, nil
}
```

`call()!` returns early on error; on success, it yields the other results.
The enclosing function must return a final `error`. Early returns preserve
named results and zero unnamed results. Defers work normally.

Details: [unions and unit](doc/tagged-unions.txt) · [error propagation](doc/bang-operator.txt).

## Install

Run `./misc/install-go++.sh` to build and install `go++`, `go++pls`, and
`go++fmt` in `~/.local/bin`. Requires a stock Go toolchain and network access.
Set `GOROOT_BOOTSTRAP` if it is not auto-detected. Re-run to update.

## Local type inference

Function signatures stay explicit; inside function bodies, generic literals,
union constructors, and calls can infer type arguments. Write `Option.Some(42)`
for `Option[int]`, or let an assignment or return supply the expected type.
Details: [local type inference](doc/local-type-inference.txt).

## Stdlib v2

`option.Option[T]` represents `Some T` or `None`. The first v2 APIs use it
for map lookups, slice searches and bounds, and HTTP status text.
`net/http/status/v2` provides union status codes.
Errors remain multiple returns with a final `error`, propagated with `!`.
Details: [stdlib v2](doc/stdlib-v2.txt).

## Musings

Many thanks to the Golang team. I love this language. I always will. It is home.
