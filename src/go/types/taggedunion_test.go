// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestTaggedUnion(t *testing.T) {
	const source = `package p
 type Foo union { Bar int; Baz string }
 type Result[T any] union { Ok T; Err string }
 func (r Result[T]) HasValue() bool {
  switch r.(variant) { case Ok: return true; case Err: return false }
 }
 var _ interface { HasValue() bool } = Result[int].Ok(42)
 type Truth bool
 func assert(r Result[int]) int {
  var truth Truth
  var value int
  value, truth = r.(Ok)
  _, _ = value, truth
  y, ok := r.(Ok)
  var _ bool = ok
  var _ int = y
  return r.(Ok)
 }
 var large Result[[64]byte]
 var makeFoo func(int) Foo = Foo.Bar
 var _ = map[Foo]int{Foo.Bar(0): 1}
 var _ = Result[int].Ok(42) == Result[int].Ok(42)
 func read(x Foo) int {
  switch data := x.(variant) {
  case Bar: return data
  case Baz: return len(data)
  }
 }`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "union.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	var formatted bytes.Buffer
	if err := format.Node(&formatted, fset, file); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(formatted.String(), "union {") || !strings.Contains(formatted.String(), ".(variant)") {
		t.Fatalf("lost union syntax: %s", &formatted)
	}
	file, err = parser.ParseFile(fset, "formatted.go", formatted.String(), 0)
	if err != nil {
		t.Fatal(err)
	}
	sizes := types.SizesFor("gc", "amd64")
	config := types.Config{Sizes: sizes}
	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue), Uses: make(map[*ast.Ident]types.Object)}
	pkg, err := config.Check("p", fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(n ast.Node) bool {
		if assertion, ok := n.(*ast.TypeAssertExpr); ok {
			if name, ok := assertion.Type.(*ast.Ident); ok && name.Name == "Ok" {
				if info.TypeOf(name) != types.Typ[types.Int] || info.Uses[name] == nil {
					t.Errorf("missing assertion variant type or use for %s", name.Name)
				}
			}
		}
		if clause, ok := n.(*ast.CaseClause); ok {
			for _, variant := range clause.List {
				if info.TypeOf(variant) == nil {
					t.Errorf("missing payload type for variant %s", variant.(*ast.Ident).Name)
				}
			}
		}
		return true
	})

	if methods := types.NewMethodSet(pkg.Scope().Lookup("Foo").Type()); methods.Len() != 0 {
		t.Fatalf("constructors leaked into method set: %v", methods)
	}
	if methods := types.NewMethodSet(pkg.Scope().Lookup("large").Type()); methods.Len() != 1 || methods.At(0).Obj().Name() != "HasValue" {
		t.Fatalf("generic union method set: %v", methods)
	}
	if size := sizes.Sizeof(pkg.Scope().Lookup("Foo").Type()); size != 32 {
		t.Fatalf("sizeof Foo = %d, want 32", size)
	}
	if size := sizes.Sizeof(pkg.Scope().Lookup("large").Type()); size != 80 {
		t.Fatalf("sizeof Result[[64]byte] = %d, want 80", size)
	}
	missing := strings.Replace(source, "case Baz: return len(data)", "", 1)
	file, err = parser.ParseFile(fset, "missing.go", missing, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = config.Check("p", fset, []*ast.File{file}, nil)
	if err == nil || !strings.Contains(err.Error(), "non-exhaustive variant switch: missing Baz") {
		t.Fatalf("missing variant diagnostic: %v", err)
	}
}

func TestTaggedUnionCases(t *testing.T) {
	const decl = `type U union { A int; B string }; `
	cases := []struct{ name, source, want string }{
		{"ordinary_union_name", `type union int; var _ = []union{1,2}; var _ = map[string]union{"x":3}`, ""},
		{"ordinary_generic_union_name", `type union[T any] int; type Alias = union[int]; var _ = []Alias{1}`, ""},
		{"ordinary_union_return", `type union int; func f() union { return 1 }`, ""},
		{"recursive_pointer", `type U union { Next *U; End int }; var _ = U.Next(nil)`, ""},
		{"recursive_value", `type U union { Next U; End int }`, "invalid recursive type"},
		{"generic_constraint", `type U[T comparable] union { A T; B string }; var _ = U[[]int]{}`, "does not satisfy comparable"},
		{"generic_key", `type U[T any] union { A T; B string }; var _ = map[U[[]int]]int{}`, "invalid map key type"},
		{"empty_union", `type U union {}`, "at least one variant"},
		{"duplicate_variant", `type U union { A int; A string }`, "redeclared"},
		{"blank_variant", `type U union { _ int }`, "one name"},
		{"wrong_constructor_arity", decl + `var _ = U.A()`, "not enough arguments"},
		{"extra_constructor_argument", decl + `var _ = U.A(1,2)`, "too many arguments"},
		{"unit_constructor_argument", `type U union { Empty }; var _ = U.Empty(unit{})`, "too many arguments"},
		{"named_empty_payload", `type Empty struct{}; type U union { Value Empty }; var _ = U.Value(Empty{})`, ""},
		{"named_empty_needs_payload", `type Empty struct{}; type U union { Value Empty }; var _ = U.Value()`, "not enough arguments"},
		{"missing_constructor", decl + `var _ = U.C(1)`, "no field or method C"},
		{"value_constructor", decl + `var u U; var _ = u.A(1)`, "no field or method A"},
		{"payload_assignment", decl + `func f(u U) { switch v := u.(variant) { case A: var _ string = v; case B: } }`, "as string"},
		{"missing_variant", decl + `func f(u U) { switch u.(variant) { case A: } }`, "missing B"},
		{"duplicate_case", decl + `func f(u U) { switch u.(variant) { case A: case A: case B: } }`, "duplicate variant A"},
		{"multiple_cases", decl + `func f(u U) { switch u.(variant) { case A,B: } }`, "exactly one variant"},
		{"qualified_case", decl + `func f(u U) { switch u.(variant) { case U.A: case B: } }`, "must be a variant name"},
		{"fallthrough", decl + `func f(u U) { switch u.(variant) { case A: fallthrough; case B: } }`, "cannot fallthrough"},
		{"blank_binding", decl + `func f(u U) { switch _ := u.(variant) { case A: case B: } }`, "no new variable"},
		{"unused_binding", decl + `func f(u U) { switch v := u.(variant) { case A: case B: } }`, "declared and not used"},
		{"case_scope", decl + `func f(u U) { switch v := u.(variant) { case A: _=v; case B: _=v }; _=v }`, "undefined: v"},
		{"default_binding", decl + `func f(u U) { switch v := u.(variant) { case A: _=v; default: var _ U = v } }`, ""},
		{"exhaustive_return", decl + `func f(u U) int { switch u.(variant) { case A: return 1; case B: return 2 } }`, ""},
		{"break_requires_return", decl + `func f(u U) int { switch u.(variant) { case A: break; case B: return 2 } }`, "missing return"},
		{"nonunion_match", `func f(x int) { switch x.(variant) { default: } }`, "not a union"},
		{"type_named_variant", `type variant int; func f(x any) variant { return x.(variant) }`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "union.go", "package p; "+tc.source, 0)
			if err == nil {
				_, err = new(types.Config).Check("p", fset, []*ast.File{file}, nil)
			}
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
}

func TestUnitAndAnchorSizes(t *testing.T) {
	const source = `package p
 type Option[T any] union { None; Some T }
 type Alias = unit
 type Units union { Implicit; Explicit unit; Alias Alias; Struct struct{} }
 var _ func() Units = Units.Implicit
 var _ func() Units = Units.Explicit
 var _ func() Units = Units.Alias
 var _ func() Units = Units.Struct
 var _ = Option[unit].Some()
 var _ = Units.Explicit()
 var integer Option[int]
 var text Option[string]
 var empty Option[unit]
 var zeroArray Option[[0]*int]
 var nested Option[Option[int]]
 var rooted Option[Option[string]]
 var u unit
 var _ struct{} = u
 func shadow() { type unit int; type U union { Empty }; var _ U = U.Empty() }
 `
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "unit.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var formatted bytes.Buffer
	if err := format.Node(&formatted, fset, f); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(formatted.String(), "None unit") || !strings.Contains(formatted.String(), "None") {
		t.Fatal("lost implicit unit syntax", formatted.String())
	}
	f, err = parser.ParseFile(fset, "formatted.go", formatted.String(), 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, arch := range []string{"amd64", "386"} {
		sizes := types.SizesFor("gc", arch)
		config := types.Config{Sizes: sizes}
		pkg, err := config.Check("p", fset, []*ast.File{f}, nil)
		if err != nil {
			t.Fatal(err)
		}
		word := sizes.Sizeof(types.Typ[types.Uintptr])
		for name, want := range map[string]int64{"integer": 2 * word, "text": 4 * word, "empty": word, "zeroArray": word, "nested": 3 * word, "rooted": 6 * word} {
			if got := sizes.Sizeof(pkg.Scope().Lookup(name).Type()); got != want {
				t.Errorf("%s %s size=%d, want %d", arch, name, got, want)
			}
		}
	}
	obj := types.Universe.Lookup("unit").(*types.TypeName)
	if !obj.IsAlias() || !types.Identical(obj.Type(), types.NewStruct(nil, nil)) {
		t.Fatal("unit alias identity")
	}
}
