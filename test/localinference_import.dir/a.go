package a

type Box[T any] struct{ Value T }
type Reversed[A, B any] = Pair[B, A]
type Pair[A, B any] struct {
	First  A
	Second B
}
type Option[T any] union {
	None
	Some T
}

func Zero[T any]() T          { var x T; return x }
func Wrap[T any](x T) Box[T]  { return Box{x} }
func Make() Box[int]          { return Box{42} }
func (b Box[T]) Copy() Box[T] { return Box{b.Value} }
