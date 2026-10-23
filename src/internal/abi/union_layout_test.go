// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi_test

import (
	"internal/abi"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

type anchorOption[T any] union {
	None
	Some T
}
type nestedScalar struct {
	A [3]int
	B unit
}
type nestedPointer struct{ A [2]string }

func TestUnionAnchorLayout(t *testing.T) {
	for _, tt := range []struct {
		typ      *abi.Type
		pointers bool
	}{
		{abi.TypeFor[anchorOption[unit]](), false},
		{abi.TypeFor[anchorOption[int]](), false},
		{abi.TypeFor[anchorOption[uintptr]](), false},
		{abi.TypeFor[anchorOption[[0]*int]](), false},
		{abi.TypeFor[anchorOption[nestedScalar]](), false},
		{abi.TypeFor[anchorOption[anchorOption[int]]](), false},
		{abi.TypeFor[anchorOption[*int]](), true},
		{abi.TypeFor[anchorOption[unsafe.Pointer]](), true},
		{abi.TypeFor[anchorOption[any]](), true},
		{abi.TypeFor[anchorOption[string]](), true},
		{abi.TypeFor[anchorOption[[]int]](), true},
		{abi.TypeFor[anchorOption[map[int]int]](), true},
		{abi.TypeFor[anchorOption[chan int]](), true},
		{abi.TypeFor[anchorOption[func()]](), true},
		{abi.TypeFor[anchorOption[[2]*int]](), true},
		{abi.TypeFor[anchorOption[nestedPointer]](), true},
		{abi.TypeFor[anchorOption[anchorOption[string]]](), true},
	} {
		t.Run(tt.typ.Kind().String(), func(t *testing.T) {
			if tt.typ.Pointers() != tt.pointers {
				t.Fatalf("Pointers()=%v, want %v", tt.typ.Pointers(), tt.pointers)
			}
			fields := tt.typ.StructType().Fields
			anchor := fields[1].Typ
			if anchor.Pointers() != tt.pointers {
				t.Fatal("anchor pointer map")
			}
			if !tt.pointers && (anchor.Size() != 0 || fields[1].Offset != fields[2].Offset) {
				t.Fatal("unused anchor occupies space")
			}
		})
	}
	if unsafe.Sizeof(anchorOption[int]{}) != 2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("scalar union size")
	}
	if unsafe.Sizeof(anchorOption[unit]{}) != unsafe.Sizeof(uintptr(0)) {
		t.Fatal("unit union size")
	}
}

func genericAnchorOption[T any](x T) anchorOption[T] { return anchorOption[T].Some(x) }

var scalarAnchorSink anchorOption[int]
var mixedAnchorSink anchorOption[string]

func TestUnionWithoutAnchor(t *testing.T) {
	if n := testing.AllocsPerRun(100, func() { scalarAnchorSink = genericAnchorOption(42) }); n != 0 {
		t.Fatalf("scalar constructor allocations: %v", n)
	}
	if n := testing.AllocsPerRun(100, func() { mixedAnchorSink = anchorOption[string].None(unit{}) }); n != 0 {
		t.Fatalf("unit constructor allocations in pointer-bearing union: %v", n)
	}
	if scalarAnchorSink != anchorOption[int].Some(42) {
		t.Fatal("scalar constructor")
	}
	if anchorOption[int].None(unit{}) != (anchorOption[int]{}) {
		t.Fatal("implicit unit")
	}
	if !reflect.TypeFor[unit]().AssignableTo(reflect.TypeFor[struct{}]()) || reflect.TypeFor[unit]() != reflect.TypeFor[struct{}]() {
		t.Fatal("unit is not an alias")
	}
	// Omission always means the builtin unit, even if its name is shadowed.
	type unit int
	type empty union{ Empty  }
	if empty.Empty(struct{}{}) != (empty{}) {
		t.Fatal("shadowed implicit unit")
	}
	if reflect.TypeFor[empty]().Variant(0).Type != reflect.TypeFor[struct{}]() {
		t.Fatal("implicit payload type")
	}
	none := reflect.ValueOf(anchorOption[int].None)
	if got := none.Call([]reflect.Value{reflect.ValueOf(struct{}{})})[0].Interface(); got != (anchorOption[int]{}) {
		t.Fatal("unit constructor function value")
	}
	value := genericAnchorOption("keep me")
	runtime.GC()
	name, payload := reflect.ValueOf(value).Variant()
	if name != "Some" || payload.String() != "keep me" {
		t.Fatal("pointer-bearing generic value")
	}
	v := reflect.New(reflect.TypeFor[anchorOption[int]]()).Elem()
	v.SetVariant("Some", reflect.ValueOf(7))
	if v.Interface() != anchorOption[int].Some(7) {
		t.Fatal("reflection constructor")
	}
	runtime.GC()
}
