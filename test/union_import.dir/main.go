// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "./a"

func read(x a.Result[int]) int {
	switch data := x.(variant) {
	case Value:
		return data
	case Error:
		panic(data)
	}
}

type Copy a.Result[int]

func forward[T any](v T) a.Result[T] { return a.Result[T].Value(v) }

func main() {
	if !a.Equal(a.Make(), a.Result[int].Value(42)) || a.Keys("key")[a.Result[string].Value("key")] != "key" {
		panic("imported generic equality and hashing")
	}
	var zero a.Result[int]
	if !a.Equal(zero, a.Result[int].Value(0)) {
		panic("imported zero equality")
	}

	var implementation interface{ HasValue() bool } = forward(45)
	if !implementation.HasValue() || read(forward(45)) != 45 {
		panic("imported generic methods")
	}
	switch data := Copy.Value(44).(variant) {
	case Value:
		if data != 44 {
			panic("derived imported union")
		}
	case Error:
		panic(data)
	}
	if read(a.Make()) != 42 || read(a.Result[int].Value(43)) != 43 {
		panic("imported union")
	}
	switch data := a.Result[int].Error("error").(variant) {
	case Value:
		panic(data)
	case Error:
		if data != "error" {
			panic("imported payload")
		}
	}
}
