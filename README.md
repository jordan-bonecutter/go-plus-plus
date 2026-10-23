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

Construct values with `Option[int].Some(42)` or `Option[int].None(unit{})`.
`unit` aliases `struct{}`; omitting a variant's type makes it a unit.

Matches must cover every variant or include a default. Extract a specific
variant with `x.(Some)` or `value, ok := x.(Some)`.

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

## Local type inference

Function signatures stay explicit. Inside function bodies, generic literals
and union constructors infer type arguments from their values:

```go
box := Box{Value: 42}     // Box[int]
option := Option.Some(42) // Option[int]
var value string = zero() // zero[string](), given func zero[T any]() T
```

Calls can also use the expected type from an assignment, argument, or return.
Ambiguous expressions still need an annotation; later statements do not
determine earlier variable types.

Details: [local type inference](doc/local-type-inference.txt).

## Musings

Many thanks to the Golang team. I love this language. I always will. It is home.
