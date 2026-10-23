// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type E int

func (E) Error() string { return "error" }

var errBoom error = E(1)

func seq(y func(int) bool) {
	for i := 0; i < 3; i++ {
		if !y(i) {
			return
		}
	}
}
func failing() error { return errBoom }
func iter(fail bool) (func(func(int) bool), error) {
	if fail {
		return seq, errBoom
	}
	return seq, nil
}
func named() (n int, err error) {
	n = 42
	defer func() { n++ }()
	for range seq {
		failing()!
	}
	return
}
func nested() (int, error) {
	for range seq {
		for range seq {
			failing()!
		}
	}
	return 7, nil
}
func header() (int, error) {
	for range iter(true)! {
	}
	return 7, nil
}
func innerHeader() (int, error) {
	for range seq {
		for range iter(true)! {
		}
	}
	return 7, nil
}
func renamed(fail bool) (int, error) {
	for range seq {
		if !fail {
			return 7, nil
		}
		break
	}
	failing()!
	return 7, nil
}
func generic[T any](v T) (T, error) {
	for range seq {
		failing()!
	}
	return v, nil
}
func main() {
	n, e := named()
	if n != 43 || e != errBoom {
		panic("named")
	}
	n, e = nested()
	if n != 0 || e != errBoom {
		panic("nested")
	}
	n, e = header()
	if n != 0 || e != errBoom {
		panic("header")
	}
	n, e = innerHeader()
	if n != 0 || e != errBoom {
		panic("inner header")
	}
	n, e = renamed(true)
	if n != 0 || e != errBoom {
		panic("renamed")
	}
	s, e := generic("x")
	if s != "" || e != errBoom {
		panic("generic")
	}
}
