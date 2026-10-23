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
