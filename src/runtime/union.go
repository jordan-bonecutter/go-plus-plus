// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/goarch"
	"unsafe"
)

// unionAnchor is the pointer-sized tag's referent. Payloads occupy overlapping
// inline storage, which the GC does not scan. This anchor retains just their
// pointer roots, using an ordinary precise Go pointer layout. Copies of a union
// share the immutable anchor. No collector or write barrier special cases are
// needed, and inactive scalar bytes can never be mistaken for pointers.
// Index must stay first: the compiler-generated tag header accesses it.
type unionAnchor struct {
	index uintptr
	typ   *_type
	roots []unsafe.Pointer
}

// unionanchor receives &payload through an interface, preserving the declared
// payload type even when it is itself an interface. It is called only by
// compiler-generated constructors. Its arguments deliberately escape: pointers
// copied to raw union storage must point to the heap, so copying a goroutine
// stack never needs to relocate raw bytes.
func unionanchor(value any, index uintptr) unsafe.Pointer {
	e := efaceOf(&value)
	t := (*ptrtype)(unsafe.Pointer(e._type)).Elem
	a := &unionAnchor{index: index, typ: t}
	if t.PtrBytes == 0 {
		return unsafe.Pointer(a)
	}
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

// A nil anchor denotes the zero value of the first variant. The marker array
// preserves that payload's type in the ordinary runtime struct descriptor.
func unionvalue(t *_type, p unsafe.Pointer) (uintptr, *_type) {
	st := (*structtype)(unsafe.Pointer(t))
	a := *(**unionAnchor)(add(p, st.Fields[1].Offset))
	if a != nil {
		return a.index, a.typ
	}
	data := (*structtype)(unsafe.Pointer(st.Fields[0].Typ))
	return 0, (*arraytype)(unsafe.Pointer(data.Fields[0].Typ)).Elem
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
