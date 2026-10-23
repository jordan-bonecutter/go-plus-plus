// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi_test

import (
	"internal/abi"
	"reflect"
	"testing"
	"unsafe"
)

type narrowUnion union {
	A int32
	B byte
}
type tinyUnion union {
	A byte
	B bool
}
type unitEnum union {
	A
	B
}

func TestUnionNaturalAlignment(t *testing.T) {
	if got := unsafe.Sizeof(narrowUnion{}); got != 8 {
		t.Fatalf("int32/byte union size=%d, want 8", got)
	}
	if got := unsafe.Alignof(narrowUnion{}); got != unsafe.Alignof(int32(0)) {
		t.Fatal("int32/byte union alignment", got)
	}
	if got := unsafe.Sizeof(tinyUnion{}); got != 2 {
		t.Fatalf("byte/bool union size=%d, want 2", got)
	}
	if got := unsafe.Sizeof(unitEnum{}); got != 1 {
		t.Fatalf("unit enum size=%d, want 1", got)
	}
	for _, pair := range [][2]reflect.Type{
		{reflect.TypeFor[anchorOption[unit]](), reflect.TypeFor[unit]()},
		{reflect.TypeFor[anchorOption[byte]](), reflect.TypeFor[byte]()},
		{reflect.TypeFor[anchorOption[int16]](), reflect.TypeFor[int16]()},
		{reflect.TypeFor[anchorOption[int32]](), reflect.TypeFor[int32]()},
		{reflect.TypeFor[anchorOption[int64]](), reflect.TypeFor[int64]()},
		{reflect.TypeFor[anchorOption[[257]byte]](), reflect.TypeFor[[257]byte]()},
		{reflect.TypeFor[anchorOption[anchorOption[byte]]](), reflect.TypeFor[anchorOption[byte]]()},
		{reflect.TypeFor[anchorOption[*int]](), reflect.TypeFor[*int]()},
		{reflect.TypeFor[anchorOption[string]](), reflect.TypeFor[string]()},
	} {
		fields := []reflect.StructField{{Name: "Payload", Type: pair[1]}}
		hasPointers := abi.TypeOf(reflect.New(pair[0]).Elem().Interface()).Pointers()
		if hasPointers {
			fields = append(fields, reflect.StructField{Name: "Anchor", Type: reflect.TypeFor[*byte]()})
		}
		fields = append(fields, reflect.StructField{Name: "Tag", Type: reflect.TypeFor[byte]()})
		layout := reflect.StructOf(fields)
		if pair[0].Size() != layout.Size() || pair[0].Align() != layout.Align() {
			t.Errorf("%v layout size/align=%d/%d, want %d/%d", pair[0], pair[0].Size(), pair[0].Align(), layout.Size(), layout.Align())
		}
	}
}

func TestTinyUnionOperations(t *testing.T) {
	x := tinyUnion.B(true)
	if x != tinyUnion.B(true) || x == tinyUnion.A(1) || map[tinyUnion]int{x: 42}[tinyUnion.B(true)] != 42 {
		t.Fatal("comparison/hash")
	}
	v := reflect.New(reflect.TypeFor[tinyUnion]()).Elem()
	v.SetVariant("B", reflect.ValueOf(true))
	if v.Interface() != x || v.IsZero() {
		t.Fatal("reflect constructor/zero")
	}
	name, p := v.Variant()
	if name != "B" || !p.Bool() {
		t.Fatal("reflect payload")
	}
	fn := reflect.ValueOf(func(x tinyUnion) tinyUnion { return x })
	if got := fn.Call([]reflect.Value{v})[0].Interface(); got != x {
		t.Fatal("reflect call")
	}
	made := reflect.MakeFunc(fn.Type(), func(args []reflect.Value) []reflect.Value { return args })
	if got := made.Interface().(func(tinyUnion) tinyUnion)(x); got != x {
		t.Fatal("reflect makefunc")
	}
	v.SetZero()
	if !v.IsZero() || v.Interface() != tinyUnion.A(0) {
		t.Fatal("zero value")
	}
}
