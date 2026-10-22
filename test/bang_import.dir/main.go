// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "./a"

func main() {
	s0, e0 := a.RangeUnnamed("value")
	if s0 != "" || e0 != a.Err {
		panic("import iterator zero")
	}
	s0, e0 = a.RangeNamed("saved")
	if s0 != "saved" || e0 != a.Err {
		panic("import iterator named")
	}

	n, e := a.Count(false)
	if n != 7 || e != nil {
		panic("import success")
	}
	n, e = a.Count(true)
	if n != 0 || e != a.Err {
		panic("import error")
	}
	s, e := a.Unnamed("value", true)
	if s != "" || e != a.Err {
		panic("import generic zero")
	}
	s, e = a.Named("saved", true)
	if s != "saved" || e != a.Err {
		panic("import generic named")
	}
}
