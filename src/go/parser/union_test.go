// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"
)

func TestUnionVariantComments(t *testing.T) {
	for _, parameters := range []string{"", "[T any]"} {
		t.Run(parameters, func(t *testing.T) {
			source := `package p
 type U` + parameters + ` union {
 // First documents the first variant.
 First int // trailing first
 // Second documents the second variant.
 Second string // trailing second
 }`
			f, err := ParseFile(token.NewFileSet(), "union.go", source, ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			u := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
			if !u.Union || len(u.Fields.List) != 2 {
				t.Fatalf("unexpected union AST: %#v", u)
			}
			for _, field := range u.Fields.List {
				if field.Doc == nil || field.Comment == nil {
					t.Errorf("variant %s lost documentation or trailing comment", field.Names[0].Name)
				}
			}
		})
	}
}

func TestUnionMissingConstraint(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "union.go", `package p
 type Option[T] union { None struct{}; Some(T) }
 var sentinel int`, AllErrors)
	if err == nil || !strings.Contains(err.Error(), "missing constraint for type parameter T (use [T any])") {
		t.Fatalf("unexpected diagnostic: %v", err)
	}
	if len(f.Decls) != 2 {
		t.Fatalf("parser did not recover after union: %d declarations", len(f.Decls))
	}
	// A type named union remains valid as an ordinary array element type.
	_, err = ParseFile(token.NewFileSet(), "array.go", `package p; type union int; const N=3; type A[N]union`, AllErrors)
	if err != nil {
		t.Fatal(err)
	}
}
