// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"internal/goarch"
	"unsafe"
)

// unionAnchor is the GC anchor pointer's referent. Payloads occupy overlapping
// inline storage, which the GC does not scan. This anchor retains just their
// pointer roots, using an ordinary precise Go pointer layout. Copies of a union
// share the immutable anchor. No collector or write barrier special cases are
// needed, and inactive scalar bytes can never be mistaken for pointers.
type unionAnchor struct {
	roots []unsafe.Pointer
}

// unionanchor receives &payload through an interface, preserving the declared
// payload type even when it is itself an interface. It is called by
// compiler-generated constructors and reflection. Its arguments deliberately escape:
// pointers copied to raw union storage must point to the heap, so copying a goroutine
// stack never needs to relocate raw bytes.
func unionanchor(value any) unsafe.Pointer {
	e := efaceOf(&value)
	t := (*ptrtype)(unsafe.Pointer(e._type)).Elem
	if t.PtrBytes == 0 {
		return nil
	}
	a := new(unionAnchor)
	mask := getGCMask(t)
	words := t.PtrBytes / goarch.PtrSize
	count := 0
	for i := uintptr(0); i < words; i++ {
		if *addb(mask, i/8)&(1<<(i%8)) != 0 {
			count++
		}
	}
	a.roots = make([]unsafe.Pointer, count)
	j := 0
	for i := uintptr(0); i < words; i++ {
		if *addb(mask, i/8)&(1<<(i%8)) != 0 {
			a.roots[j] = *(*unsafe.Pointer)(add(e.data, i*goarch.PtrSize))
			j++
		}
	}
	KeepAlive(value)
	return unsafe.Pointer(a)
}

// The inline tag determines the active variant, including for the zero value.
// Marker arrays preserve payload types in the runtime struct descriptor.
func unionvalue(t *_type, p unsafe.Pointer) (uintptr, *_type) {
	st := (*structtype)(unsafe.Pointer(t))
	index := uintptr(abi.UnionIndex(st, p))
	data := (*structtype)(unsafe.Pointer(st.Fields[0].Typ))
	return index, (*arraytype)(unsafe.Pointer(data.Fields[index].Typ)).Elem
}

func unionequal(t *_type, p, q unsafe.Pointer) bool {
	i, payload := unionvalue(t, p)
	j, _ := unionvalue(t, q)
	return i == j && payload.Equal(p, q)
}

func unionhash(t *_type, p unsafe.Pointer, seed uintptr) uintptr {
	i, payload := unionvalue(t, p)
	seed = memhash(unsafe.Pointer(&i), seed, goarch.PtrSize)
	return typehash(payload, p, seed)
}

//go:linkname reflect_unionAnchor reflect.unionAnchor
func reflect_unionAnchor(payload any) unsafe.Pointer {
	return unionanchor(payload)
}
