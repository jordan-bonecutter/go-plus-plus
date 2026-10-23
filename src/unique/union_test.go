// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unique_test

import (
	"runtime"
	"strings"
	"testing"
	"unique"
)

type unionKey union {
	Text   string
	Number int
}

func TestUnion(t *testing.T) {
	x := unionKey.Text(strings.Repeat("x", 1024))
	h := unique.Make(x)
	runtime.GC()
	if h.Value() != x || unique.Make(x) != h {
		t.Fatal("union handle equality")
	}
	if unique.Make(unionKey.Number(0)) == unique.Make(unionKey.Text("")) {
		t.Fatal("variant identity")
	}
	a := unique.Make([1]unionKey{x})
	runtime.GC()
	if a.Value()[0] != x {
		t.Fatal("union in array")
	}
}
