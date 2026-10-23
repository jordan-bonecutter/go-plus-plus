// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi

import "unsafe"

// UnionTagSize returns the smallest tag width in bytes that can distinguish
// count variants. The first variant has index zero.
func UnionTagSize(count uint64) uintptr {
	switch {
	case count <= 1<<8:
		return 1
	case count <= 1<<16:
		return 2
	case count <= 1<<32:
		return 4
	default:
		return 8
	}
}

// IsTaggedUnion reports whether s uses the compiler's private union layout:
// overlapping payload storage, GC anchor pointer, and inline unsigned tag.
func (s *StructType) IsTaggedUnion() bool {
	return len(s.Fields) == 3 && s.Fields[0].Name.Name() == "union·data" && s.Fields[1].Name.Name() == "union·anchor" && s.Fields[2].Name.Name() == "union·tag"
}

// UnionIndex reads an inline union tag without dereferencing the anchor.
// s must describe a tagged union.
func UnionIndex(s *StructType, p unsafe.Pointer) uint64 {
	tag := unsafe.Add(p, s.Fields[2].Offset)
	switch s.Fields[2].Typ.Size_ {
	case 1:
		return uint64(*(*uint8)(tag))
	case 2:
		return uint64(*(*uint16)(tag))
	case 4:
		return uint64(*(*uint32)(tag))
	case 8:
		return *(*uint64)(tag)
	default:
		panic("invalid union tag size")
	}
}
