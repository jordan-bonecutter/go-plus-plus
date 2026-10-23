package maps_test

import (
	"maps/v2"
	"testing"
)

func TestLookup(t *testing.T) {
	m := map[string]int{"zero": 0}
	if maps.Get(m, "zero").OrPanic() != 0 || !maps.Get(m, "missing").IsNone() {
		t.Fatal(m)
	}
	if !maps.Take(m, "zero").IsSome() || len(m) != 0 {
		t.Fatal(m)
	}
	if !maps.Take(map[string]int(nil), "zero").IsNone() {
		t.Fatal("nil map")
	}
}
