// run

package main

type Box[T any] struct{ Value T }
type List[T any] []T
type Table[K comparable, V any] map[K]V
type Option[T any] union {
	None
	Some T
}

func zero[T any]() T             { var x T; return x }
func identity[T any](x T) T      { return x }
func empty[T any]() (T, error)   { return zero[T](), nil }
func returnString() string       { return zero() }
func propagate() (string, error) { return empty()!, nil }
func takeInt(x int) int          { return x }

func main() {
	b := Box{Value: "hello"}
	if b.Value != "hello" {
		panic("box inference")
	}
	xs := List{1, 2, 3}
	if xs[2] != 3 {
		panic("slice inference")
	}
	m := Table{"one": 1}
	if m["one"] != 1 {
		panic("map inference")
	}
	var nested Box[Box[int]] = Box{Box{42}}
	if nested.Value.Value != 42 {
		panic("nested inference")
	}
	var x string = zero()
	x = zero()
	if x != "" || returnString() != "" || takeInt(zero()) != 0 {
		panic("result inference")
	}
	if s, err := propagate(); s != "" || err != nil {
		panic("bang inference")
	}
	var id func(int) int = identity
	if id(42) != 42 {
		panic("generic function value")
	}
	o := Option.Some(42)
	if value := o.(Some); value != 42 {
		panic("union inference")
	}
	var empty func() Option[string] = Option.None
	var none Option[string] = Option.None()
	if empty() != none {
		panic("unit constructor inference")
	}
	switch none.(variant) {
	case None:
	case Some:
		panic("empty union inference")
	}
	var some func(int) Option[int] = Option.Some
	if value := some(7).(Some); value != 7 {
		panic("constructor function inference")
	}
}
