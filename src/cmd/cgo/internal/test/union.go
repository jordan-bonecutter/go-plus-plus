// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgotest

// static void unionKindPointer(void *p) {}
import "C"

import (
	"runtime"
	"testing"
	"unsafe"
)

type cgoUnion union {
	Number  int
	Pointer *int
}

func testUnionKind(t *testing.T) {
	// Scalar payloads do not expose the implementation's GC anchor to C.
	scalar := cgoUnion.Number(7)
	C.unionKindPointer(unsafe.Pointer(&scalar))
	p := new(int)
	value := cgoUnion.Pointer(p)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("accepted unpinned union payload pointer")
			}
		}()
		C.unionKindPointer(unsafe.Pointer(&value))
	}()
	var pin runtime.Pinner
	pin.Pin(p)
	defer pin.Unpin()
	C.unionKindPointer(unsafe.Pointer(&value))
}
