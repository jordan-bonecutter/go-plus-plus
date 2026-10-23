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

func TestBang(t *testing.T) {
	prefix := `package p
 func f() (int,error) { return 1,nil }
 func g() (int,string,error) { return 1,"",nil }
 func e() error { return nil }
 func plain() int { return 1 }
 `
	cases := []struct{ body, want string }{
		{"func test() bool { return !\nfalse }", ""},
		{"func test() bool { return ! /*\ncomment*/ false }", ""},
		{"func test() bool { return ! !\ntrue }", ""},
		{`func test() (int,error) { x:=f()!; return x,nil }`, ""},
		{`func test() error { e()!
 return nil }`, ""},
		{`func test() (x int,err error) { x=f()!; return }`, ""},
		{`func test() error { x,s:=g()!; _,_=x,s; return nil }`, ""},
		{`func take(int,string){}; func test() error { take(g()!); return nil }`, ""},
		{`func test() error { if !true {e()!}; return nil }`, ""},
		{`func test() { e()! }`, "enclosing function"},
		{`var x=f()!`, "enclosing function"},
		{`func test() error { plain()!; return nil }`, "call to return a final error"},
		{`func test() error { int(1)!; return nil }`, "error-returning function call"},
		{`func test() error { len("")!; return nil }`, "error-returning function call"},
		{`func test() error { go e()!; return nil }`, "not allowed in go call"},
		{`func test() error { defer e()!; return nil }`, "not allowed in defer call"},
		{`func test() error { _=func(){e()!}; return nil }`, "enclosing function"},
		{`func test() error { var x int; _=x!; return nil }`, "postfix ! requires a function call"},
	}
	for _, tt := range cases {
		t.Run(tt.body, func(t *testing.T) {
			source := prefix + tt.body
			fs := token.NewFileSet()
			f, err := parser.ParseFile(fs, "bang.go", source, parser.AllErrors)
			if err == nil {
				_, err = new(types.Config).Check("p", fs, []*ast.File{f}, nil)
			}
			if tt.want != "" {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("got %v, want %q", err, tt.want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var buf bytes.Buffer
			if err := format.Node(&buf, fs, f); err != nil {
				t.Fatal(err)
			}
			fs = token.NewFileSet()
			f, err = parser.ParseFile(fs, "formatted.go", buf.Bytes(), 0)
			if err != nil {
				t.Fatalf("%s: %v", buf.String(), err)
			}
			if _, err = new(types.Config).Check("p", fs, []*ast.File{f}, nil); err != nil {
				t.Fatal(err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok && call.Bang.IsValid() {
					if call.End() != call.Bang+1 {
						t.Error("call End omits !")
					}
					if !strings.HasSuffix(types.ExprString(call), "!") {
						t.Error("expression string omits !")
					}
				}
				return true
			})
		})
	}
}

// FuzzBang checks that malformed postfix syntax never crashes the front end and
// that formatting a parsed file preserves whether it type-checks.
func FuzzBang(f *testing.F) {
	for _, source := range []string{
		`package p; func e() error{return nil}; func f() error {e()!;return nil}`,
		`package p; func e()(int,error){return 1,nil}; func f()(int,error){return e()!,nil}`,
		`package p; func e()(int,error){return 1,nil}; func f()(x int,err error){x=e()!;return}`,
		`package p; func e()(int,string,error){return 1,"",nil}; func f()error{x,s:=e()!;_,_=x,s;return nil}`,
		`package p; func e()error{return nil}; func f(){e()!}`,
		`package p; func e()error{return nil}; func f()error{e()!!;return nil}`,
		`package p; func e()error{return nil}; var x=e()!`,
		"package p\nfunc e()error{return nil}\nfunc f()error{e()! // comment\nreturn nil}",
		`package p; func e()error{return nil}; func f()error{defer e()!;return nil}`,
		`package p; type U union {Value int}; func e()(U,error){return U.Value(1),nil}; func f()(U,error){return e()!,nil}`,
	} {
		f.Add(source)
	}
	f.Fuzz(func(t *testing.T, source string) {
		if len(source) > 4096 || !strings.Contains(source, "!") {
			return
		}
		fs := token.NewFileSet()
		file, err := parser.ParseFile(fs, "bang.go", source, parser.AllErrors)
		if err != nil {
			return
		}
		_, before := new(types.Config).Check("p", fs, []*ast.File{file}, nil)
		var buf bytes.Buffer
		if err := format.Node(&buf, fs, file); err != nil {
			t.Fatal(err)
		}
		fs = token.NewFileSet()
		formatted, err := parser.ParseFile(fs, "formatted.go", buf.Bytes(), parser.AllErrors)
		if err != nil {
			t.Fatalf("format produced invalid syntax: %v\n%s", err, buf.String())
		}
		_, after := new(types.Config).Check("p", fs, []*ast.File{formatted}, nil)
		if (before == nil) != (after == nil) {
			t.Fatalf("format changed type checking: before=%v after=%v\n%s", before, after, buf.String())
		}
	})
}
