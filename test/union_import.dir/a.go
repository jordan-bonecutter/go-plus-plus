// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

type Result[T any] union {
	Value T
	Error string
}

func Make() Result[int] { return Result[int].Value(42) }

func (r Result[T]) HasValue() bool {
	switch r.(variant) {
	case Value:
		return true
	case Error:
		return false
	}
}

func Equal[T comparable](x, y Result[T]) bool { return x == y }
func Keys[T comparable](value T) map[Result[T]]T {
	return map[Result[T]]T{Result[T].Value(value): value}
}

// Private variants retain their declaration's package through derived types.
type Private union {
	hidden int
	Public string
}

func Hidden() Private { return Private.hidden(7) }
