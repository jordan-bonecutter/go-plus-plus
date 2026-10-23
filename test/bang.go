// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type bangError string

func (e bangError) Error() string { return string(e) }

var boom error = bangError("boom")
var calls int

func value(b bool) (int, error) {
	calls++
	if b {
		return 99, boom
	}
	return 7, nil
}
func pair(b bool) (int, string, error) {
	if b {
		return 99, "bad", boom
	}
	return 7, "ok", nil
}
func only(b bool) error {
	if b {
		return boom
	}
	return nil
}
func unnamed(b bool) (int, string, *int, error) {
	x := value(b)!
	return x, "ok", &x, nil
}
func named(b bool) (n int, s string, err error) {
	n, s, err = 42, "saved", bangError("old")
	value(b)!
	return
}
func shadowed() (n int, err error) {
	n = 42
	{
		n := 88
		_ = n
		only(true)!
	}
	return
}
func deferred(b bool) (n int, err error) {
	n = 42
	defer func() {
		n++
		if b && err != boom {
			panic("defer error")
		}
	}()
	only(b)!
	return
}
func zeros() (n int, err error)            { n = 42; return unnamedFailure() }
func unnamedFailure() (int, error)         { x := 42; _ = x; only(true)!; return 1, nil }
func multiple(b bool) (int, string, error) { n, s := pair(b)!; return n, s, nil }
func add(n int, s string) int              { return n + len(s) }
func arguments(b bool) (int, error)        { return add(pair(b)!), nil }
func genericValue[T any](v T, b bool) (T, error) {
	if b {
		return v, boom
	}
	return v, nil
}
func generic[T any](v T, b bool) (out T, err error) { out = v; genericValue(v, b)!; return }
func nested(b bool) error {
	f := func() (int, error) { return value(b)!, nil }
	f()!
	return nil
}
func boolean(b bool) (bool, error) {
	if b {
		return true, boom
	}
	return true, nil
}
func shortCircuit(b bool) (bool, error) { return false && boolean(b)!, nil }

var order []int

func step(n int) int              { order = append(order, n); return n }
func ordered(b bool) (int, error) { return step(1) + value(b)! + step(2), nil }
func loop(b bool) error {
	for i := 0; i < 2; i++ {
		only(b)!
	}
	return nil
}

type receiver struct{}

func (receiver) Read(b bool) (int, error) { return value(b) }
func methods(b bool) (int, error)         { var r receiver; return r.Read(b)!, nil }
func interfaceMethod(b bool) (int, error) {
	var r interface{ Read(bool) (int, error) } = receiver{}
	return r.Read(b)!, nil
}
func variadic(xs ...int) (int, error) { return len(xs), nil }
func dots() (int, error)              { xs := []int{1, 2}; return variadic(xs...)!, nil }

type errorAlias = error

func aliasCall() errorAlias                           { return boom }
func alias() errorAlias                               { aliasCall()!; return nil }
func blank() (_ int, err error)                       { only(true)!; return 99, nil }
func zeroGeneric[T any]() (T, error)                  { only(true)!; var zero T; return zero, nil }
func containers() (map[int]int, []int, func(), error) { only(true)!; return nil, nil, nil, nil }

type typedNilError struct{}

func (*typedNilError) Error() string { return "nil pointer" }
func typedNilCall() error            { var e *typedNilError; return e }
func typedNil() error                { typedNilCall()!; panic("typed nil must propagate") }
func main() {
	if alias() != boom {
		panic("alias")
	}
	z, err := blank()
	if z != 0 || err != boom {
		panic("blank result")
	}
	zs, err := zeroGeneric[string]()
	if zs != "" || err != boom {
		panic("generic zero")
	}
	m, sl, fn, err := containers()
	if m != nil || sl != nil || fn != nil || err != boom {
		panic("container zeros")
	}
	if typedNil() == nil {
		panic("typed nil error")
	}

	n, s, p, e := unnamed(false)
	if n != 7 || s != "ok" || p == nil || e != nil {
		panic("success")
	}
	n, s, p, e = unnamed(true)
	if n != 0 || s != "" || p != nil || e != boom {
		panic("unnamed zeros")
	}
	n, s, e = named(true)
	if n != 42 || s != "saved" || e != boom {
		panic("named")
	}
	n, s, e = named(false)
	if n != 42 || s != "saved" || e == nil || e == boom {
		panic("named success preserves error")
	}
	n, e = shadowed()
	if n != 42 || e != boom {
		panic("shadowed")
	}
	n, e = deferred(true)
	if n != 43 || e != boom {
		panic("deferred")
	}
	n, e = zeros()
	if n != 0 || e != boom {
		panic("inline return")
	}
	n, s, e = multiple(false)
	if n != 7 || s != "ok" || e != nil {
		panic("tuple")
	}
	n, s, e = multiple(true)
	if n != 0 || s != "" || e != boom {
		panic("tuple failure")
	}
	n, e = arguments(false)
	if n != 9 || e != nil {
		panic("arguments")
	}
	n, e = arguments(true)
	if n != 0 || e != boom {
		panic("arguments failure")
	}
	n, e = generic(12, true)
	if n != 12 || e != boom {
		panic("generic named")
	}
	if nested(true) != boom || nested(false) != nil {
		panic("closure")
	}
	b, e := shortCircuit(true)
	if b || e != nil {
		panic("short circuit")
	}
	calls = 0
	order = nil
	n, e = ordered(true)
	if n != 0 || e != boom || calls != 1 || len(order) != 1 || order[0] != 1 {
		panic("failure order")
	}
	order = nil
	n, e = ordered(false)
	if n != 10 || e != nil || len(order) != 2 {
		panic("success order")
	}
	if loop(true) != boom || loop(false) != nil {
		panic("loop")
	}
	n, e = methods(false)
	if n != 7 || e != nil {
		panic("method")
	}
	n, e = interfaceMethod(true)
	if n != 0 || e != boom {
		panic("interface")
	}
	n, e = dots()
	if n != 2 || e != nil {
		panic("variadic")
	}
}
