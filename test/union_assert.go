// run

package main

type Option[T any] union {
	None  struct{}
	Some  T
	Other T
}

var calls int

func next() Option[int] {
	calls++
	return Option[int].Some(42)
}

func shadowed() int {
	bool, true, false, panic := 1, 2, 3, 4
	_, _, _, _ = bool, true, false, panic
	x := Option[int].Some(7)
	y, ok := x.(Some)
	if ok {
		return y
	}
	return x.(Some)
}

func generic[T any](x Option[T]) T {
	y, ok := x.(Some)
	if !ok {
		panic("generic assertion")
	}
	return y
}

func nested() int {
	return func() Option[int] {
		switch y := Option[int].Some(19).(variant) {
		case Some:
			return Option[int].Some(y)
		case None:
			return Option[int]{}
		case Other:
			return Option[int].Other(y)
		}
	}().(Some)
}

func genericBool[B ~bool](x Option[int]) B {
	var y int
	var ok B
	y, ok = x.(Some)
	_ = y
	return ok
}

func main() {
	if !genericBool[bool](Option[int].Some(1)) || genericBool[bool](Option[int]{}) {
		panic("generic bool")
	}

	// A comma-ok assertion must not recheck and lower its operand's body.
	if y, ok := func() Option[int] {
		switch y := Option[int].Some(23).(variant) {
		case Some:
			return Option[int].Some(y)
		case None:
			return Option[int]{}
		case Other:
			return Option[int].Other(y)
		}
	}().(Some); !ok || y != 23 {
		panic("nested comma-ok")
	}
	type Truth bool
	var truth Truth
	var n int
	n, truth = Option[int].Some(31).(Some)
	if !truth || n != 31 {
		panic("named boolean assignment")
	}

	if nested() != 19 || generic(Option[string].Some("yes")) != "yes" {
		panic("nested or generic")
	}
	var assigned int
	var matched bool
	assigned, matched = Option[int].Some(8).(Some)
	if assigned != 8 || !matched {
		panic("assignment")
	}
	var declared, valid = Option[int].Some(9).(Some)
	if declared != 9 || !valid {
		panic("var declaration")
	}
	type Inner union {
		Int  int
		Text string
	}
	outer := Option[Inner].Some(Inner.Int(12))
	if outer.(Some).(Int) != 12 {
		panic("chained")
	}
	type Alias = Option[int]
	type Derived Option[int]
	if Alias.Some(4).(Some) != 4 || Derived.Some(5).(Some) != 5 {
		panic("alias/derived")
	}
	var dynamic Option[any] = Option[any].Some(nil)
	if payload, ok := dynamic.(Some); !ok || payload != nil {
		panic("nil interface")
	}
	if payload, ok := dynamic.(Other); ok || payload != nil {
		panic("interface zero")
	}

	if shadowed() != 7 {
		panic("shadowed names")
	}
	var x Option[int]
	if y, ok := x.(Some); ok || y != 0 {
		panic("zero mismatch")
	}
	x = Option[int].Some(42)
	if y, ok := x.(Some); !ok || y != 42 {
		panic("match")
	}
	if y, ok := x.(Other); ok || y != 0 {
		panic("same payload, different variant")
	}
	if x.(Some) != 42 {
		panic("single assertion")
	}
	if y, ok := (next().(Some)); !ok || y != 42 || calls != 1 {
		panic("evaluation")
	}
	p := Option[*int].Some(new(int))
	if y, ok := p.(Some); !ok || y == nil {
		panic("pointer")
	}
	if y, ok := p.(Other); ok || y != nil {
		panic("pointer mismatch")
	}
	defer func() {
		if recover() == nil {
			panic("missing assertion panic")
		}
	}()
	_ = x.(None)
}
