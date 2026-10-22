// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type E int

func (E) Error() string { return "err" }

var boom error = E(1)

func slicePair(b bool) ([]int, int, error) {
	if b {
		return nil, 0, boom
	}
	return []int{1}, 2, nil
}
func sliceCopy(b bool) ([]int, []int, error) {
	if b {
		return nil, nil, boom
	}
	return make([]int, 1), []int{7}, nil
}
func mapKey(m map[int]int, b bool) (map[int]int, int, error) {
	if b {
		return nil, 0, boom
	}
	return m, 7, nil
}
func intPair(b bool) (int, int, error) {
	if b {
		return 0, 0, boom
	}
	return 7, 8, nil
}
func take(xs ...int) int {
	n := 0
	for _, v := range xs {
		n += v
	}
	return n
}
func appendPair(b bool) ([]int, error)       { return append(slicePair(b)!), nil }
func copyPair(b bool) (int, error)           { return copy(sliceCopy(b)!), nil }
func deletePair(m map[int]int, b bool) error { delete(mapKey(m, b)!); return nil }
func variadic(b bool) (int, error)           { return take(intPair(b)!), nil }
func printer(b bool) error {
	if b {
		println(intPair(b)!)
	}
	return nil
}
func passthrough() (int, error, error) { return 7, nil, nil }
func tail() (int, error)               { return passthrough()! }
func main() {
	for _, b := range []bool{false, true} {
		s, e := appendPair(b)
		if b {
			if s != nil || e != boom {
				panic("append failure")
			}
		} else if len(s) != 2 || s[1] != 2 || e != nil {
			panic("append success")
		}
		n, e := copyPair(b)
		if b {
			if n != 0 || e != boom {
				panic("copy failure")
			}
		} else if n != 1 || e != nil {
			panic("copy success")
		}
		m := map[int]int{7: 9}
		e = deletePair(m, b)
		if b {
			if m[7] != 9 || e != boom {
				panic("delete failure")
			}
		} else if len(m) != 0 || e != nil {
			panic("delete success")
		}
		n, e = variadic(b)
		if b {
			if n != 0 || e != boom {
				panic("variadic failure")
			}
		} else if n != 15 || e != nil {
			panic("variadic success")
		}
	}
	if printer(true) != boom {
		panic("printer failure")
	}
	n, e := tail()
	if n != 7 || e != nil {
		panic("return tuple")
	}
}
