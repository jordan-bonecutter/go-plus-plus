// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi_test

import (
	"internal/abi"
	"testing"
	"unsafe"
)

func TestUnionTagSize(t *testing.T) {
	for _, tt := range []struct {
		count uint64
		size  uintptr
	}{
		{1, 1}, {256, 1}, {257, 2}, {65536, 2}, {65537, 4},
		{1 << 32, 4}, {1<<32 + 1, 8}, {1<<64 - 1, 8},
	} {
		if got := abi.UnionTagSize(tt.count); got != tt.size {
			t.Errorf("UnionTagSize(%d) = %d, want %d", tt.count, got, tt.size)
		}
	}
}

func TestUnionIndex(t *testing.T) {
	for _, x := range []any{uint8(255), uint16(65535), uint32(1<<32 - 1), uint64(1<<64 - 1)} {
		var ptr unsafe.Pointer
		var want uint64
		switch x := x.(type) {
		case uint8:
			ptr, want = unsafe.Pointer(&x), uint64(x)
		case uint16:
			ptr, want = unsafe.Pointer(&x), uint64(x)
		case uint32:
			ptr, want = unsafe.Pointer(&x), uint64(x)
		case uint64:
			ptr, want = unsafe.Pointer(&x), x
		}
		st := &abi.StructType{Fields: make([]abi.StructField, 3)}
		st.Fields[2].Typ = abi.TypeOf(x)
		if got := abi.UnionIndex(st, ptr); got != want {
			t.Errorf("tag %T: got %d, want %d", x, got, want)
		}
	}
}

type kindUnion union {
	Number int
	Text   string
}

func TestUnionDescriptor(t *testing.T) {
	typ := abi.TypeFor[kindUnion]()
	if typ.Kind() != abi.Union || typ.Kind().String() != "union" {
		t.Fatalf("kind %v", typ.Kind())
	}
	st := typ.StructType()
	if st == nil || !st.IsTaggedUnion() {
		t.Fatal("missing union layout")
	}
	if abi.Union >= 32 {
		t.Fatal("union kind does not fit reflection's five-bit kind field")
	}
	for _, width := range []int{4, 8} {
		if abi.RTypeSize(abi.Union, width) != abi.RTypeSize(abi.Struct, width) {
			t.Fatal("descriptor size")
		}
	}
	// The physical tag remains private even though the public kind is distinct.
	if st.Fields[2].Typ.Kind() != abi.Uint8 {
		t.Fatal("tag width")
	}
	if typ.DescriptorSize() <= abi.RTypeSize(abi.Union, int(unsafe.Sizeof(uintptr(0)))) {
		t.Fatal("descriptor fields missing")
	}
}
