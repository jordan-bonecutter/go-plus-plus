// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"runtime"
	"unsafe"
)

type Foo union {
	Bar   int
	Baz   string
	Other int
}

type Pointer union {
	Number uintptr
	Value  *int
}

type Dynamic union {
	Data interface{}
	Unit struct{}
}

type Empty union{ Unit struct{} }

func read(x Foo) string {
	switch data := x.(variant) {
	case Bar:
		var y int = data
		if y != 10 {
			panic("Bar payload")
		}
		return "bar"
	case Baz:
		var z string = data
		return z
	case Other:
		return "other"
	}
}

//go:noinline
func grow(n int) {
	var b [1024]byte
	b[0] = byte(n)
	if n > 0 {
		grow(n - 1)
	}
	runtime.KeepAlive(b)
}

func local() {
	new, max, byte, uintptr, nil := 1, 2, 3, 4, 5
	_ = new + max + byte + uintptr + nil
	type __union_data int
	type value union {
		Data __union_data
		Unit struct{}
	}
	type derived value
	type alias = value
	var makeValue func(__union_data) derived = derived.Data
	switch data := makeValue(42).(variant) {
	case Data:
		if data != 42 {
			panic("local constructor")
		}
	case Unit:
		panic("local variant")
	}
	switch data := alias.Data(43).(variant) {
	case Data:
		if data != 43 {
			panic("local alias")
		}
	case Unit:
		panic("alias variant")
	}
}

type Result[T any] union {
	Ok  T
	Err string
}

type hasValue interface{ HasValue() bool }

func (r Result[T]) HasValue() bool {
	switch r.(variant) {
	case Ok:
		return true
	case Err:
		return false
	}
}

func (r Result[T]) Get() T {
	switch data := r.(variant) {
	case Ok:
		return data
	case Err:
		panic(data)
	}
}

func resultSize[T any](v Result[T]) uintptr { return unsafe.Sizeof(v) }

func makeResult[T any](v T) Result[T] { return Result[T].Ok(v) }

func generic() {
	var i hasValue = makeResult(42)
	if !i.HasValue() {
		panic("generic union interface")
	}
	i = Result[int].Err("no value")
	if i.HasValue() {
		panic("generic error variant")
	}
	if makeResult(42).Get() != 42 {
		panic("generic constructor")
	}
	var large [64]byte
	large[63] = 42
	value := makeResult(large)
	if value.Get()[63] != 42 {
		panic("generic large payload")
	}
	if unsafe.Sizeof(value) != 64+unsafe.Sizeof(uintptr(0)) || reflect.TypeOf(value).Size() != unsafe.Sizeof(value) || resultSize(value) != unsafe.Sizeof(value) {
		panic("generic layout")
	}
	pointer := new(int)
	*pointer = 43
	rooted := makeResult(pointer)
	pointer = nil
	runtime.GC()
	grow(32)
	if *rooted.Get() != 43 {
		panic("generic pointer roots")
	}
}

func comparableUnion() {
	var zero Foo
	if zero != Foo.Bar(0) || Foo.Bar(10) != Foo.Bar(10) || Foo.Bar(10) == Foo.Other(10) {
		panic("variant equality")
	}
	s := string([]byte{'h', 'i'})
	keys := map[Foo]int{zero: 1, Foo.Baz(s): 2, Foo.Other(0): 3}
	if keys[Foo.Bar(0)] != 1 || keys[Foo.Baz("hi")] != 2 || keys[Foo.Other(0)] != 3 {
		panic("union map keys")
	}
	if any(Foo.Baz(s)) != any(Foo.Baz("hi")) || Result[int].Ok(42) != Result[int].Ok(42) {
		panic("boxed/generic equality")
	}
	type Nested struct{ Value [2]Foo }
	if (Nested{[2]Foo{Foo.Bar(1), Foo.Baz(s)}}) != (Nested{[2]Foo{Foo.Bar(1), Foo.Baz("hi")}}) {
		panic("nested equality")
	}
	type Outer union {
		Inner Foo
		Unit  struct{}
	}
	boxed := map[any]int{Foo.Baz(s): 1, Outer.Inner(Foo.Baz(s)): 2}
	if boxed[Foo.Baz("hi")] != 1 || boxed[Outer.Inner(Foo.Baz("hi"))] != 2 {
		panic("dynamic union hash")
	}
	if !reflect.ValueOf(Foo.Baz(s)).Equal(reflect.ValueOf(Foo.Baz("hi"))) || reflect.ValueOf(Dynamic.Data([]int{1})).Comparable() {
		panic("reflect union equality")
	}
	type Float union{ Value float64 }
	var negativeZero float64
	negativeZero = -negativeZero
	floats := map[Float]int{Float.Value(negativeZero): 1}
	if floats[Float.Value(0)] != 1 {
		panic("float union hash")
	}
	for _, operation := range []func(){
		func() { x := Dynamic.Data([]int{1}); _ = x == x },
		func() { _ = map[Dynamic]int{Dynamic.Data([]int{1}): 1} },
	} {
		func() {
			defer func() {
				if recover() == nil {
					panic("interface payload must panic")
				}
			}()
			operation()
		}()
	}
	if Empty.Unit(struct{}{}) != (Empty{}) {
		panic("empty equality")
	}
}

func main() {
	comparableUnion()
	generic()
	local()
	// Repeated payload types must remain distinct variants, and sizeof must not
	// grow when a smaller variant is added. Tags occupy one machine word.
	if unsafe.Sizeof(Foo{}) != unsafe.Sizeof(uintptr(0))+unsafe.Sizeof("") {
		panic("union layout")
	}
	if unsafe.Sizeof(Pointer{}) != 2*unsafe.Sizeof(uintptr(0)) {
		panic("pointer union layout")
	}
	if unsafe.Sizeof(Empty{}) != unsafe.Sizeof(uintptr(0)) {
		panic("empty payload layout")
	}
	var construct func(int) Foo = Foo.Bar
	if read(construct(10)) != "bar" || read(Foo.Baz("baz")) != "baz" || read(Foo.Other(10)) != "other" {
		panic("dispatch")
	}
	var zero Pointer
	switch data := zero.(variant) {
	case Number:
		if data != 0 {
			panic("zero value")
		}
	case Value:
		panic("zero variant")
	}
	calls := 0
	input := func() Foo { calls++; return Foo.Bar(10) }
	switch data := input().(variant) {
	case Bar:
		if data != 10 {
			panic("input")
		}
	default:
		panic("default")
	}
	if calls != 1 {
		panic("repeated input")
	}
	// Make the inline bytes the only user-visible reference, then collect and
	// grow the stack. Both direct-interface pointer payloads and string payloads
	// need their anchors to retain the underlying allocations.
	p := new(int)
	*p = 42
	values := []Pointer{Pointer.Value(p), Pointer.Number(^uintptr(0))}
	p = nil
	dynamic := Dynamic.Data([2]int{19, 23})
	str := Foo.Baz(string([]byte{'a', 'b', 'c'}))
	for i := 0; i < 3; i++ {
		runtime.GC()
		grow(32)
	}
	switch data := values[0].(variant) {
	case Value:
		if *data != 42 {
			panic("collected pointer")
		}
	case Number:
		panic("pointer tag")
	}
	switch data := values[1].(variant) {
	case Number:
		if data != ^uintptr(0) {
			panic("scalar bits")
		}
	case Value:
		panic("scalar tag")
	}
	switch data := dynamic.(variant) {
	case Data:
		values := data.([2]int)
		if values[0]+values[1] != 42 {
			panic("collected interface payload")
		}
	case Unit:
		panic("interface tag")
	}

	if read(str) != "abc" {
		panic("collected string")
	}
}
