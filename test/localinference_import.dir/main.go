package main

import "./a"

func main() {
	b := a.Box{42}
	if b.Value != 42 || b.Copy().Value != 42 || a.Make().Value != 42 {
		panic("imported literals")
	}
	var p a.Reversed[string, int] = a.Reversed{}
	if p.First != 0 || p.Second != "" {
		panic("imported alias")
	}
	var x string = a.Zero()
	if x != "" || a.Wrap(42).Value != 42 {
		panic("imported functions")
	}
	o := a.Option.Some(42)
	if value := o.(Some); value != 42 {
		panic("imported union")
	}
	var none a.Option[string] = a.Option.None(unit{})
	switch none.(variant) {
	case None:
	case Some:
		panic("imported empty union")
	}
}
