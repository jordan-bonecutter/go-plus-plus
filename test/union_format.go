// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"reflect"
	"runtime"
)

type Foo union {
	Bar int
	Baz string
}
type Generic[T any] union {
	Value T
	Empty struct{}
}
type Pretty union{ Number int }

func (p Pretty) String() string { return "custom" }

type Literal union{ Number int }

func (p Literal) GoString() string { return "literal" }

func main() {
	type Private union {
		hidden int
		Public int
	}
	privateName, privatePayload := reflect.ValueOf(Private.hidden(7)).Variant()
	if privateName != "hidden" || privatePayload.CanInterface() {
		panic("private variant exposed")
	}
	type Unicode union{ Étage int }
	_, unicodePayload := reflect.ValueOf(Unicode.Étage(8)).Variant()
	if !unicodePayload.CanInterface() {
		panic("exported Unicode variant hidden")
	}

	var zero Foo
	for _, tc := range []struct {
		format string
		value  any
		want   string
	}{
		{"%v", zero, "Foo.Bar(0)"},
		{"%+v", Foo.Bar(42), "Foo.Bar(42)"},
		{"%#v", Foo.Baz("hello"), `main.Foo.Baz("hello")`},
		{"%v", Foo.Baz("hello"), "Foo.Baz(hello)"},
		{"%v", []Foo{zero, Foo.Bar(1)}, "[Foo.Bar(0) Foo.Bar(1)]"},
		{"%v", &zero, "&Foo.Bar(0)"},
		{"%v", Generic[int].Value(7), "Generic[int].Value(7)"},
		{"%v", Generic[Foo].Value(zero), "Generic[main.Foo].Value(Foo.Bar(0))"},
		{"%v", Pretty.Number(1), "custom"},
		{"%#v", Literal.Number(1), "literal"},
		{"%v", struct{ hidden Foo }{zero}, "{Foo.Bar(0)}"},
	} {
		if got := fmt.Sprintf(tc.format, tc.value); got != tc.want {
			panic(fmt.Sprintf("format %s: got %q, want %q", tc.format, got, tc.want))
		}
	}
	if got := fmt.Sprintln(zero); got != "Foo.Bar(0)\n" {
		panic(got)
	}
	name, payload := reflect.ValueOf(Foo.Baz("retained")).Variant()
	runtime.GC()
	if name != "Baz" || payload.String() != "retained" || payload.CanSet() {
		panic("variant reflection")
	}
	hidden := reflect.ValueOf(struct{ hidden Foo }{Foo.Bar(3)}).Field(0)
	_, payload = hidden.Variant()
	if payload.CanInterface() {
		panic("unexported payload exposed")
	}
	for _, value := range []reflect.Value{{}, reflect.ValueOf(1), reflect.ValueOf(struct{}{})} {
		name, payload := value.Variant()
		if name != "" || payload.IsValid() {
			panic("nonunion Variant")
		}
	}
}
