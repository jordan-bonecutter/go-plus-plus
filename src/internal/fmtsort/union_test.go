// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fmtsort_test

import (
	"internal/fmtsort"
	"reflect"
	"testing"
)

type unionKey union {
	Z      string
	A      string
	hidden int
}

func TestUnionSort(t *testing.T) {
	want := []unionKey{unionKey.Z("a"), unionKey.Z("b"), unionKey.A("a"), unionKey.hidden(1), unionKey.hidden(2)}
	m := make(map[unionKey]int)
	for i, key := range want {
		m[key] = i
	}
	for range 10 {
		sorted := fmtsort.Sort(reflect.ValueOf(m))
		for i, entry := range sorted {
			if entry.Key.Interface().(unionKey) != want[i] {
				t.Fatalf("index %d: got %v, want %v", i, entry.Key, want[i])
			}
		}
	}
}
