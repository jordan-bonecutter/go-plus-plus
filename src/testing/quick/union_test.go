// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quick

import (
	"math/rand"
	"reflect"
	"testing"
)

type unionChoice union {
	Number int
	Text   string
	hidden int
}
type unionHidden union{ hidden int }

func TestUnionValue(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	seen := make(map[string]bool)
	for range 100 {
		v, ok := Value(reflect.TypeFor[unionChoice](), r)
		if !ok {
			t.Fatal("could not generate union")
		}
		name, payload := v.Variant()
		if !payload.CanInterface() {
			t.Fatal("generated hidden variant")
		}
		seen[name] = true
	}
	if !seen["Number"] || !seen["Text"] {
		t.Fatalf("did not generate both variants: %v", seen)
	}
	if _, ok := Value(reflect.TypeFor[unionHidden](), r); ok {
		t.Fatal("generated hidden variant")
	}
	if err := Check(func(v unionChoice) bool { return reflect.TypeOf(v).IsUnion() }, nil); err != nil {
		t.Fatal(err)
	}
}
