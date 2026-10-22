// errorcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type Foo union {
	Bar int
	Baz string
}

var _ = Foo.Bar("wrong") // ERROR "cannot use .* as int"

func missing(x Foo) {
	switch x.(variant) { // ERROR "non-exhaustive variant switch: missing Baz"
	case Bar:
	}
}

func cases(x Foo) {
	switch x.(variant) {
	case Bar:
	case Bar: // ERROR "duplicate variant Bar"
	case Baz:
	case Nope: // ERROR "unknown variant Nope"
	}
}

func payload(x Foo) {
	switch data := x.(variant) {
	case Bar:
		var _ string = data // ERROR "cannot use data .* as string"
	case Baz:
		var _ int = data // ERROR "cannot use data .* as int"
	}
}

func ordinary(x int) {
	switch x.(variant) { // ERROR "not a union"
	}
}

type NonComparable union {
	Slice  []int
	Number int
}

var _ = map[NonComparable]int{}  // ERROR "invalid map key type"
func compareBad(x NonComparable) { _ = x == x } // ERROR "cannot be compared"

func arity() {
	_ = Foo.Bar()      // ERROR "not enough arguments"
	_ = Foo.Bar(1, 2)  // ERROR "too many arguments"
	_ = Foo.Missing(1) // ERROR "no field or method Missing"
	var f Foo
	_ = f.Bar(1) // ERROR "no field or method Bar"
}

func invalidCases(x Foo) {
	switch x.(variant) { // ERROR "missing Bar, Baz"
	case Bar, Baz: // ERROR "exactly one variant"
	}
	switch x.(variant) { // ERROR "missing Bar"
	case Foo.Bar: // ERROR "must be a variant name"
	case Baz:
	}
	switch x.(variant) {
	case Bar:
		fallthrough // ERROR "cannot fallthrough"
	case Baz:
	}
}

func invalidBindings(x Foo) {
	switch _ := x.(variant) { // ERROR "no new variable"
	case Bar:
	case Baz:
	}
	switch unused := x.(variant) { // ERROR "declared and not used"
	case Bar:
	case Baz:
	}
	switch v := x.(variant) {
	case Bar:
		_ = v
	case Baz:
		_ = v
	}
	_ = v // ERROR "undefined: v"
}

func brokenReturn(x Foo) int {
	switch x.(variant) {
	case Bar:
		break
	case Baz:
		return 1
	}
} // ERROR "missing return"

type Constrained[T comparable] union {
	Value T
	Empty struct{}
}

var _ Constrained[[]int] // ERROR "does not satisfy comparable"
type Generic[T any] union {
	Value T
	Empty struct{}
}

var _ map[Generic[[]int]]int // ERROR "invalid map key type"
type Recursive union { // ERROR "invalid recursive type"
	Value Recursive
	Empty struct{}
}
