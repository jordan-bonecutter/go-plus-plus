// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type UnitAlias = unit
type Units union {
	Implicit
	Explicit unit
	Alias    UnitAlias
	Struct   struct{}
	Value    int
}

func main() {
	constructors := []func() Units{Units.Implicit, Units.Explicit, Units.Alias, Units.Struct}
	for i, construct := range constructors {
		v := construct()
		for j, other := range constructors {
			if (v == other()) != (i == j) {
				panic("unit variant identity")
			}
		}
	}
	if Units.Explicit().(Explicit) != (unit{}) || Units.Value(42).(Value) != 42 {
		panic("variant payload")
	}
	type unit int
	type Local union{ Empty  }
	var makeLocal func() Local = Local.Empty
	if makeLocal() != (Local{}) {
		panic("shadowed unit")
	}
}
