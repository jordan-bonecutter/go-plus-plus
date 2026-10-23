// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package a

type Option[T any] union {
	None
	Some T
}

func Make[T any](x T) Option[T] { return Option[T].Some(x) }
func Read[T any](x Option[T]) T {
	switch p := x.(variant) {
	case Some:
		return p
	case None:
		var zero T
		return zero
	}
}
func Nothing() Option[int] { return Option[int].None() }
