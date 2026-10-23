// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"unsafe"
)

type E string

func (e E) Error() string { return string(e) }

var boom error = E("boom")

func val[T any](v T, b bool) (T, error) {
	if b {
		return v, boom
	}
	return v, nil
}
func only(b bool) error {
	if b {
		return boom
	}
	return nil
}
func seq(y func(int) bool) {
	for i := 0; i < 3; i++ {
		if !y(i) {
			return
		}
	}
}

var trace []int

func mark(i int) { trace = append(trace, i) }
func deferredArg(b bool) (n int, e error) {
	n = 42
	defer func() { mark(1) }()
	defer func(x int) { mark(x) }(val(2, b)!)
	mark(3)
	return
}
func launchArg(b bool) (int, error) {
	ch := make(chan int, 1)
	go func(x int) { ch <- x }(val(7, b)!)
	return <-ch, nil
}
func recoverDefer() (n int, e error) {
	n = 42
	defer func() {
		if e != boom {
			panic("error not set")
		}
		panic("caught")
	}()
	only(true)!
	return
}
func selectLHS(b bool) (int, error) {
	ch := make(chan int, 1)
	ch <- 7
	xs := []int{1}
	select {
	case xs[val(0, b)!] = <-ch:
	}
	return xs[0], nil
}
func selectUnchosen() (int, error) {
	var ch chan int
	xs := []int{1}
	select {
	case xs[val(0, true)!] = <-ch:
	default:
	}
	return xs[0], nil
}
func selectOperand(b bool) (int, error) {
	ch := make(chan int, 1)
	ch <- 7
	select {
	case x := <-val(ch, b)!:
		return x, nil
	}
}
func initLoop(b bool) (int, error) {
	for i := val(0, b)!; i < 2; i++ {
	}
	return 7, nil
}
func postLoop(b bool) (int, error) {
	i := 0
	for ; i < 2; only(b)! {
		i++
	}
	return i, nil
}
func short(b bool) (bool, error)                { return true || val(false, b)!, nil }
func expression(b bool) (int, error)            { return val(2, b)! + val(3, false)!*val(4, false)!, nil }
func funcResult(b bool) (int, error)            { return val(func() (int, error) { return 7, nil }, b)!()!, nil }
func sizeof() (uintptr, error)                  { return unsafe.Sizeof(val(7, true)!), nil }
func tuple() (int, string, error)               { return 7, "ok", nil }
func tupleVar() (int, error)                    { var n, s = tuple()!; return n + len(s), nil }
func pass(n int, s string) (int, string, error) { return n, s, nil }
func tupleRet() (int, string, error)            { return pass(tuple()!) }
func emptyReturn() (int, error)                 { only(false)!; return 7, nil }
func iteratorDefer() (n int, e error) {
	n = 42
	for range seq {
		defer func() { n++ }()
		only(true)!
	}
	return
}
func iteratorClosure() (int, error) {
	for range seq {
		_, e := func() (int, error) { only(true)!; return 7, nil }()
		if e == boom {
			return 8, nil
		}
	}
	return 9, nil
}
func arrayZero() ([3]int, struct{ P *int }, error) {
	only(true)!
	return [3]int{1}, struct{ P *int }{}, nil
}
func assertEq(n int, e error, want int, wantErr error) {
	if n != want || e != wantErr {
		panic(fmt.Sprintf("got (%v,%v), want (%v,%v)", n, e, want, wantErr))
	}
}
func extra() (int, error, error) { return 7, nil, nil }
func tail() (int, error)         { return extra()! }
func main() {
	n0, e0 := tail()
	assertEq(n0, e0, 7, nil)

	n, e := deferredArg(true)
	assertEq(n, e, 42, boom)
	if fmt.Sprint(trace) != "[1]" {
		panic(trace)
	}
	trace = nil
	n, e = deferredArg(false)
	assertEq(n, e, 42, nil)
	if fmt.Sprint(trace) != "[3 2 1]" {
		panic(trace)
	}
	for _, b := range []bool{false, true} {
		want, err := 7, error(nil)
		if b {
			want, err = 0, boom
		}
		n, e = launchArg(b)
		assertEq(n, e, want, err)
		n, e = selectLHS(b)
		assertEq(n, e, want, err)
		n, e = selectOperand(b)
		assertEq(n, e, want, err)
		n, e = initLoop(b)
		assertEq(n, e, want, err)
		n, e = funcResult(b)
		assertEq(n, e, want, err)
	}
	n, e = selectUnchosen()
	assertEq(n, e, 1, nil)
	n, e = postLoop(true)
	assertEq(n, e, 0, boom)
	b, e := short(true)
	if !b || e != nil {
		panic("or")
	}
	n, e = expression(false)
	assertEq(n, e, 14, nil)
	n, e = expression(true)
	assertEq(n, e, 0, boom)
	sz, e := sizeof()
	if sz != unsafe.Sizeof(7) || e != nil {
		panic("sizeof")
	}
	n, e = tupleVar()
	assertEq(n, e, 9, nil)
	n, s, e := tupleRet()
	if n != 7 || s != "ok" || e != nil {
		panic("return tuple")
	}
	n, e = iteratorDefer()
	assertEq(n, e, 43, boom)
	n, e = iteratorClosure()
	assertEq(n, e, 8, nil)
	a, st, e := arrayZero()
	if a != [3]int{} || st.P != nil || e != boom {
		panic("array zeros")
	}
	func() {
		defer func() {
			if recover() != "caught" {
				panic("recover defer")
			}
		}()
		recoverDefer()
	}()
}
