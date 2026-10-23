// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"./a"
	"reflect"
	"runtime"
	"unsafe"
)

func main() {
	x := a.Make(42)
	y := a.Make("hello")
	runtime.GC()
	if a.Read(x) != 42 || a.Read(y) != "hello" || a.Read(a.Nothing()) != 0 {
		panic("imported generic union")
	}
	if unsafe.Sizeof(x) != 2*unsafe.Sizeof(uintptr(0)) || unsafe.Sizeof(y) != 4*unsafe.Sizeof(uintptr(0)) {
		panic("imported union layout")
	}
	v := reflect.New(reflect.TypeOf(x)).Elem()
	v.SetVariant("Some", reflect.ValueOf(7))
	if v.Interface() != a.Make(7) {
		panic("imported reflection")
	}
}
