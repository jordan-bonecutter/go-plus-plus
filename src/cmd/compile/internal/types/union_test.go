// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"cmd/internal/src"
	"testing"
)

func TestUnionHasPointers(t *testing.T) {
	oldPtr, oldReg, oldMax, oldByte := PtrSize, RegSize, MaxWidth, Types[TUINT8]
	t.Cleanup(func() { PtrSize, RegSize, MaxWidth, Types[TUINT8] = oldPtr, oldReg, oldMax, oldByte })
	PtrSize, RegSize, MaxWidth = 8, 8, 1<<50
	scalar := newType(TUINT8)
	Types[TUINT8] = scalar
	pkg := NewPkg("union-layout-test", "test")
	field := func(name string, typ *Type) *Field { return NewField(src.NoXPos, pkg.Lookup(name), typ) }
	union := func(payload *Type) *Type {
		data := NewStruct([]*Field{field("union·variant·Value", NewArray(payload, 0)), field("_", NewStruct(nil)), field("union·raw", NewArray(scalar, 0))})
		return NewStruct([]*Field{field("union·data", data), field("union·anchor", NewPtr(NewStruct(nil))), field("union·tag", scalar)})
	}
	for _, tt := range []struct {
		payload *Type
		want    bool
	}{
		{scalar, false},
		{NewArray(scalar, 3), false},
		{NewStruct([]*Field{field("X", scalar)}), false},
		{NewPtr(scalar), true},
		{NewInterface(nil), true},
		{NewArray(NewPtr(scalar), 0), false},
		{NewArray(NewPtr(scalar), 1), true},
		{union(scalar), false},
		{union(NewPtr(scalar)), true},
	} {
		typ := union(tt.payload)
		if got := typ.HasPointers(); got != tt.want {
			t.Errorf("union of %v: HasPointers=%v, want %v", tt.payload, got, tt.want)
		}
		if got := typ.Field(1).Type.HasPointers(); got != tt.want {
			t.Error("anchor layout disagrees with HasPointers")
		}
		if !tt.want && typ.Field(1).Type.Size() != 0 {
			t.Error("pointer-free union has an anchor slot")
		}
	}
}
