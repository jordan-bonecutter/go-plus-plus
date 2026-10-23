package option_test

import (
	"option"
	"testing"
)

func TestOption(t *testing.T) {
	var none option.Option[int]
	if !none.IsNone() || none.UnwrapOr(7) != 7 {
		t.Fatal(none)
	}
	some := option.Option.Some(0)
	if !some.IsSome() || some == none {
		t.Fatal(some)
	}
	calls := 0
	fallback := func() int { calls++; return 9 }
	if some.UnwrapOrElse(fallback) != 0 || calls != 0 {
		t.Fatal(calls)
	}
	if none.UnwrapOrElse(fallback) != 9 || calls != 1 {
		t.Fatal(calls)
	}
}

func TestCombinators(t *testing.T) {
	some := option.Option.Some(3)
	if option.Map(some, func(v int) string { return "yes" }).OrPanic() != "yes" {
		t.Fatal(some)
	}
	if option.AndThen(some, func(v int) option.Option[int] { return option.Option.Some(v * 2) }).OrPanic() != 6 {
		t.Fatal(some)
	}
	if !some.Filter(func(v int) bool { return v > 4 }).IsNone() {
		t.Fatal(some)
	}
	if option.Flatten(option.Option.Some(some)).OrPanic() != 3 {
		t.Fatal(some)
	}
	if !option.Flatten(option.Option[option.Option[int]].None()).IsNone() {
		t.Fatal("flatten")
	}
	called := false
	option.Map(option.Option[int].None(), func(v int) int { called = true; return v })
	if called {
		t.Fatal("mapped None")
	}
}

func TestOrPanicNone(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	option.Option[int].None().OrPanic()
}

func TestOrElse(t *testing.T) {
	calls := 0
	fallback := func() option.Option[int] {
		calls++
		return option.Option.Some(9)
	}
	some := option.Option.Some(0)
	if some.OrElse(fallback) != some || calls != 0 {
		t.Fatal(calls)
	}
	if option.Option[int].None().OrElse(fallback).OrPanic() != 9 || calls != 1 {
		t.Fatal(calls)
	}
	if !option.Option[int].None().OrElse(func() option.Option[int] { return option.Option[int].None() }).IsNone() {
		t.Fatal("None fallback")
	}
}

func TestAndThenNone(t *testing.T) {
	called := false
	bind := func(v int) option.Option[string] {
		called = true
		return option.Option[string].None()
	}
	if !option.AndThen(option.Option[int].None(), bind).IsNone() || called {
		t.Fatal("bound None")
	}
	if !option.AndThen(option.Option.Some(1), bind).IsNone() || !called {
		t.Fatal("bound Some")
	}
}
