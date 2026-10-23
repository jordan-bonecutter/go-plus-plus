// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"
	"reflect"
	"runtime"
)

type Uncomparable union {
	Data  []int
	Empty struct{}
}

type Sum[T comparable] union {
	First  T
	Second T
}

func (s Sum[T]) Index() int {
	switch s.(variant) {
	case First:
		return 0
	case Second:
		return 1
	}
}

func comparison(f func() bool) (equal, failed bool) {
	defer func() {
		if recover() != nil {
			failed = true
		}
	}()
	return f(), false
}

// Compare every pair through the language, both interface representations,
// reflection, and map lookup against the payload's own Go semantics.
func properties[T comparable](payloads []T) {
	constructors := []func(T) Sum[T]{Sum[T].First, Sum[T].Second}
	for ai, constructA := range constructors {
		for _, av := range payloads {
			a := constructA(av)
			if a.Index() != ai {
				panic("method receiver")
			}
			if reflect.ValueOf(a).Comparable() != reflect.ValueOf(&av).Elem().Comparable() {
				panic("Comparable disagrees with payload")
			}
			if reflect.ValueOf(a).IsZero() != (ai == 0 && reflect.ValueOf(&av).Elem().IsZero()) {
				panic("IsZero disagrees with payload")
			}
			for bi, constructB := range constructors {
				for _, bv := range payloads {
					b := constructB(bv)
					want, shouldPanic := comparison(func() bool { return ai == bi && av == bv })
					var ia, ib interface{ Index() int } = a, b
					for _, test := range []func() bool{
						func() bool { return a == b },
						func() bool { return any(a) == any(b) },
						func() bool { return ia == ib },
						func() bool { return reflect.ValueOf(a).Equal(reflect.ValueOf(b)) },
					} {
						equal, failed := comparison(test)
						if equal != want || failed != shouldPanic {
							panic("equality disagrees with payload")
						}
					}
					if reflect.DeepEqual(a, b) != (ai == bi && reflect.DeepEqual(av, bv)) {
						panic("DeepEqual disagrees with payload")
					}
					// Unhashable dynamic payloads must panic even in an empty map.
					if !reflect.ValueOf(&bv).Elem().Comparable() {
						_, failed := comparison(func() bool { return map[Sum[T]]int{}[b] == 0 })
						if !failed {
							panic("unhashable empty lookup")
						}
						continue
					}
					if !reflect.ValueOf(&av).Elem().Comparable() {
						continue
					}
					if (map[Sum[T]]bool{a: true}[b]) != want {
						panic("static map lookup")
					}
					if (map[any]bool{a: true}[b]) != want {
						panic("interface map lookup")
					}
					// Obtain the runtime hash independently of the compiler-generated hasher.
					typ := reflect.MapOf(reflect.TypeOf(a), reflect.TypeOf(uint16(0)))
					m := reflect.MakeMap(typ)
					m.SetMapIndex(reflect.ValueOf(a), reflect.ValueOf(uint16(1)))
					if m.MapIndex(reflect.ValueOf(b)).IsValid() != want {
						panic("runtime map lookup")
					}
				}
			}
		}
	}
	// The implicit zero value must be interchangeable with an explicit first
	// variant zero payload, including through generic boxing.
	var z T
	a, b := Sum[T]{}, Sum[T].First(z)
	if a != b || any(a) != any(b) || map[Sum[T]]bool{a: true}[b] != true {
		panic("implicit zero")
	}
}

func main() {
	properties([]int{-1, 0, 1, 42})
	properties([]string{"", "a", "same", string([]byte("same"))})
	properties([]struct{}{{}})
	properties([][0]*int{{}})
	properties([][2]string{{}, {"a", "b"}, {"a", "b"}})
	properties([]float64{0, math.Copysign(0, -1), 1, math.Inf(1), math.NaN()})
	properties([]complex128{0, complex(math.Copysign(0, -1), 0), 1 + 2i, complex(math.NaN(), 0)})
	p, q := new(int), new(int)
	properties([]*int{nil, p, p, q})
	type Record struct {
		Name   string
		Number int
	}
	properties([]Record{{}, {"a", 1}, {"a", 1}, {"b", 2}})
	properties([]Sum[any]{Sum[any]{}, Sum[any].First(42), Sum[any].Second(42), Sum[any].First([]int{1})})
	properties([]any{nil, (*int)(nil), 0, int64(0), "", []int{1}, map[int]int{}, func() {}, [0]Uncomparable{}, Uncomparable.Empty()})
	runtime.GC()
}
