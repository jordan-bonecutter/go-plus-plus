// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

type failure string

func (e failure) Error() string { return string(e) }

var Err error = failure("boom")

func Value[T any](v T, fail bool) (T, error) {
	if fail {
		return v, Err
	}
	return v, nil
}
func Unnamed[T any](v T, fail bool) (T, error)       { x := Value(v, fail)!; return x, nil }
func Named[T any](v T, fail bool) (out T, err error) { out = v; Value(v, fail)!; return }
func Count(fail bool) (int, error)                   { return Value(7, fail)!, nil }

func seq(y func(int) bool) { y(1) }
func RangeUnnamed[T any](v T) (T, error) {
	for range seq {
		Value(v, true)!
	}
	return v, nil
}
func RangeNamed[T any](v T) (out T, err error) {
	out = v
	for range seq {
		Value(v, true)!
	}
	return
}
