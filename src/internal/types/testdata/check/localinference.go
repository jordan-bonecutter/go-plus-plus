package localinference

type Box[T any] struct{ Value T }
type Pair[A, B any] struct {
	First  A
	Second B
}
type List[T any] []T
type Table[K comparable, V any] map[K]V
type Fixed[T any] [2]T
type Number[T ~int] struct{ Value T }
type Alias[T any] = Box[T]
type Reverse[A, B any] = Pair[B, A]
type Wrapped[T any] = Box[[]T]

func zero[T any]() T                { var result T; return result }
func pair[A, B any](a A) Pair[A, B] { return Pair[A, B]{First: a} }
func identity[T any](x T) T         { return x }
func takeBox(x Box[int])            {}
func takeInt(x int)                 {}

func literals() {
	var _ Box[int] = Box{42}
	b := Box{Value: "hello"}
	var _ Box[string] = b
	var _ Pair[int, string] = Pair{First: 1, Second: "two"}
	var _ List[int] = List{1, 2, 3}
	var _ Table[string, int] = Table{"one": 1}
	var _ Fixed[int] = Fixed{1, 2}
	var _ Alias[int] = Alias{42}
	var _ Reverse[string, int] = Reverse{}
	var _ Wrapped[int] = Wrapped{}
	var _ Box[Box[int]] = Box{Box{42}}
	var _ Box[func(int) int] = Box{identity}
	var _ Box[string] = Box{}
	takeBox(Box{1})
	var _ Number[int] = Number{1}
	_ = Number /* ERROR "does not satisfy" */ {"bad"}
	_ = Box /* ERROR "cannot infer" */ {}
	_ = Box /* ERROR "cannot infer" */ {nil}
	_ = Pair{First: 1, Second: 2}
	_ = List /* ERROR "mismatched types" */ {1, "bad"}
	_ = Box{func() { var unused /* ERROR "declared and not used" */ int }}
}

func calls() int {
	var _ string = zero()
	takeInt(zero())
	var _ Pair[int, string] = pair(1)
	var _ Pair[int, string] = pair[int](1)
	var _ int = identity(1)
	var x int
	x = zero()
	_ = x
	var _ string = identity /* ERROR "cannot use" */ (1)
	_ = zero /* ERROR "cannot infer" */ ()
	return zero()
}

// Explicit function signatures remain required, even for private functions.
var _ = Box /* ERROR "without instantiation" */ {1}

type Option[T any] union {
	None
	Some T
}

func unions() {
	x := Option.Some(42)
	var _ Option[int] = x
	var _ Option[string] = Option.None()
	var empty func() Option[string] = Option.None
	var _ Option[string] = empty()
	var _ Option[string] = Option.None(unit /* ERROR "too many arguments" */ {})
	var constructor func(int) Option[int] = Option.Some
	var _ Option[int] = constructor(7)
	_ = Option /* ERROR "cannot infer" */ .None()
	_ = Option.Some() /* ERROR "not enough arguments" */
	_ = Option.Some(1, 2 /* ERROR "too many arguments" */)
}

func explicit(x Box /* ERROR "without instantiation" */) {}
