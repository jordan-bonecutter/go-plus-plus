// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"runtime"
)

type E int

func (E) Error() string { return "boom" }

var boom error = E(1)

type Payload struct {
	Text    string
	Padding [128]*int
}
type U[T any] union {
	Value T
	Empty struct{}
}

func fail() error { return boom }
func value[T any](x T, b bool) (T, error) {
	if b {
		return x, boom
	}
	return x, nil
}
func load(b bool) (U[*Payload], error) {
	p := &Payload{Text: "hello"}
	return value(U[*Payload].Value(p), b)
}
func unnamed(b bool) (U[*Payload], error) { return load(b)!, nil }
func named(b bool) (out U[*Payload], err error) {
	out = U[*Payload].Value(&Payload{Text: "saved"})
	defer runtime.GC()
	load(b)!
	return
}
func read(u U[*Payload]) string {
	switch p := u.(variant) {
	case Value:
		return p.Text
	case Empty:
		return "empty"
	}
	panic("unreachable")
}
func generic[T any](x T) (U[T], error)                     { fail()!; return U[T].Value(x), nil }
func genericCall[F ~func() (int, error)](f F) (int, error) { return f()!, nil }
func genericTarget() (int, error)                          { return 7, boom }

var initialized = func() error { fail()!; panic("not returned") }()

func main() {
	for i := 0; i < 100; i++ {
		u, e := unnamed(false)
		if e != nil || read(u) != "hello" {
			panic("success")
		}
		u, e = unnamed(true)
		var z U[*Payload]
		if e != boom || !reflect.DeepEqual(u, z) {
			panic("zero")
		}
		u, e = named(true)
		runtime.GC()
		if e != boom || read(u) != "saved" {
			panic("named")
		}
	}
	g, e := generic([256]int{})
	var z U[[256]int]
	if e != boom || !reflect.DeepEqual(g, z) {
		panic("generic")
	}
	n, e := genericCall(genericTarget)
	if n != 0 || e != boom {
		panic("function constraint")
	}
	if initialized != boom {
		panic("initializer")
	}
}
