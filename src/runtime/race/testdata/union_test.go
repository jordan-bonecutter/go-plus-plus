// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package race_test

import (
	"internal/abi"
	"reflect"
	"testing"
	"unsafe"
)

type raceUnion union{ Data [32]int }

func TestRaceReflectUnionRead(t *testing.T) {
	var x raceUnion
	tag := (*byte)(unsafe.Add(unsafe.Pointer(&x), abi.TypeFor[raceUnion]().StructType().Fields[2].Offset))
	done := make(chan bool, 1)
	go func() { *tag = 0; done <- true }()
	_ = reflect.ValueOf(&x).Elem().Interface()
	<-done
}

func TestRaceReflectUnionWrite(t *testing.T) {
	var x raceUnion
	tag := (*byte)(unsafe.Add(unsafe.Pointer(&x), abi.TypeFor[raceUnion]().StructType().Fields[2].Offset))
	done := make(chan bool, 1)
	go func() { *tag = 0; done <- true }()
	reflect.ValueOf(&x).Elem().Set(reflect.ValueOf(raceUnion.Data([32]int{1})))
	<-done
}

func TestRaceReflectUnionZero(t *testing.T) {
	var x raceUnion
	tag := (*byte)(unsafe.Add(unsafe.Pointer(&x), abi.TypeFor[raceUnion]().StructType().Fields[2].Offset))
	done := make(chan bool, 1)
	go func() { *tag = 0; done <- true }()
	reflect.ValueOf(&x).Elem().SetZero()
	<-done
}
