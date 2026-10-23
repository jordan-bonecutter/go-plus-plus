package slices_test

import (
	"slices/v2"
	"testing"
)

func TestOptionalValues(t *testing.T) {
	s := []int{0, 4, 2}
	if slices.First(s).OrPanic() != 0 || slices.Last(s).OrPanic() != 2 {
		t.Fatal(s)
	}
	if !slices.Get(s, -1).IsNone() || !slices.Get(s, len(s)).IsNone() {
		t.Fatal(s)
	}
	if slices.Index(s, 0).OrPanic() != 0 || !slices.Index(s, 8).IsNone() {
		t.Fatal(s)
	}
	if slices.Find(s, func(v int) bool { return v > 1 }).OrPanic() != 4 {
		t.Fatal(s)
	}
	if slices.Min(s).OrPanic() != 0 || slices.Max(s).OrPanic() != 4 {
		t.Fatal(s)
	}
	if !slices.Min([]int(nil)).IsNone() || !slices.First([]int{}).IsNone() {
		t.Fatal("empty")
	}
}
