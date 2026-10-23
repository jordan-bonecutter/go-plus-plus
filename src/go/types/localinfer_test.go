// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

// Compare equal-sized function bodies using explicit and inferred instantiations.
// Parse outside the timed loop to isolate type checking from parsing and codegen.
func BenchmarkLocalInference(b *testing.B) {
	for _, inferred := range []bool{false, true} {
		name := "explicit"
		literal, call := "Box[int]{42}", "zero[int]()"
		if inferred {
			name, literal, call = "inferred", "Box{42}", "zero()"
		}
		b.Run(name, func(b *testing.B) {
			var src strings.Builder
			src.WriteString("package p; type Box[T any] struct{Value T}; func zero[T any]() T {var x T; return x}; func f(){\n")
			for range 200 {
				fmt.Fprintf(&src, "_ = %s; var _ int = %s\n", literal, call)
			}
			src.WriteString("}")
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "bench.go", src.String(), parser.SkipObjectResolution)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				var conf types.Config
				if _, err := conf.Check("p", fset, []*ast.File{file}, nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestLocalInferenceAdversarial(t *testing.T) {
	const decl = `package p
 type Box[T any] struct { Value T }
 type Option[T any] union { None; Some T }
 type Pair[A, B any] struct { A A; B B }
 type Swapped[A, B any] = Pair[B, A]
 func zero[T any]() T { var x T; return x }
 func checked[T any]() (T, error) { return zero[T](), nil }
 `
	for _, test := range []struct{ name, body, want string }{
		{"nil_payload", `func f() Option[*int] { return Option.Some(nil) }`, ""},
		{"pointer_literal", `func f() *Box[string] { return &Box{} }`, ""},
		{"alias", `func f() Swapped[int, string] { return Swapped{A: "s", B: 1} }`, ""},
		{"nested", `func f() Box[Box[int]] { return Box{Value: Box{Value: 7}} }`, ""},
		{"bang_result", `func f() (string, error) { return checked()!, nil }`, ""},
		{"unit_constructor", `func f() Option[string] { return Option.None() }`, ""},
		{"ambiguous", `func f() { _ = Box{} }`, "cannot infer"},
		{"unit_argument", `func f() Option[string] { return Option.None(unit{}) }`, "too many arguments"},
		{"wrong_result", `func f() Box[string] { return Box{Value: 7} }`, "cannot use"},
		{"duplicate_field", `func f() { _ = Box{Value: 1, Value: 2} }`, "duplicate field"},
		{"unknown_field", `func f() { _ = Box{Missing: 1} }`, "cannot infer"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fs := token.NewFileSet()
			file, err := parser.ParseFile(fs, "adversarial.go", decl+test.body, 0)
			if err != nil {
				t.Fatal(err)
			}
			var config types.Config
			_, err = config.Check("p", fs, []*ast.File{file}, nil)
			if test.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v, want %q", err, test.want)
			}
		})
	}
}
