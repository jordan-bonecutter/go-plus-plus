package slices

import (
	"cmp"
	"option"
	"slices"
)

func Get[S ~[]E, E any](s S, index int) option.Option[E] {
	if index < 0 || index >= len(s) {
		return option.Option[E].None()
	}
	return option.Option.Some(s[index])
}

func First[S ~[]E, E any](s S) option.Option[E] { return Get(s, 0) }
func Last[S ~[]E, E any](s S) option.Option[E]  { return Get(s, len(s)-1) }

func Index[S ~[]E, E comparable](s S, value E) option.Option[int] {
	index := slices.Index(s, value)
	return option.From(index, index >= 0)
}

func IndexFunc[S ~[]E, E any](s S, f func(E) bool) option.Option[int] {
	index := slices.IndexFunc(s, f)
	return option.From(index, index >= 0)
}

func Find[S ~[]E, E any](s S, f func(E) bool) option.Option[E] {
	return option.Map(IndexFunc(s, f), func(index int) E { return s[index] })
}

func Min[S ~[]E, E cmp.Ordered](s S) option.Option[E] {
	if len(s) == 0 {
		return option.Option[E].None()
	}
	return option.Option.Some(slices.Min(s))
}

func Max[S ~[]E, E cmp.Ordered](s S) option.Option[E] {
	if len(s) == 0 {
		return option.Option[E].None()
	}
	return option.Option.Some(slices.Max(s))
}
