// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Run from the patched gopls module; see build-gopls-unions.sh.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/interp"
	"golang.org/x/tools/go/ssa/ssautil"
	"honnef.co/go/tools/go/ir"
	"honnef.co/go/tools/go/ir/irutil"
)

const source = `package main
 type E int
 func(E)Error()string{return "boom"}
 var boom error=E(1)
 func value(fail bool)(int,error){if fail{return 99,boom};return 7,nil}
 func pair()(int,string,error){return 7,"ok",nil}
 func only()error{return boom}
 func take(n int,s string)int{return n+len(s)}
 func seq(y func(int)bool){y(1)}
 func unnamed(fail bool)(int,error){return value(fail)!,nil}
 func named()(n int,e error){n=42;defer func(){n++}();value(true)!;return}
 func tuple()(int,error){n,s:=pair()!;if n!=7||s!="ok"{panic("tuple")};return take(pair()!),nil}
 func iterator()(n int,e error){n=42;for range seq{only()!};return}
 func discard()error{pair()!;only()!;return nil}
 func extra()(int,error,error){return 7,nil,nil}
 func tail()(int,error){return extra()!}
 type U union{ A int; B string }
 func unionValue()(U,error){return U.A(7),nil}
 func unionBang()(U,error){return unionValue()!,nil}
 func genericValue[T any](x T,b bool)(T,error){if b{return x,boom};return x,nil}
 func generic[T any](x T,b bool)(T,error){return genericValue(x,b)!,nil}
 func arrays()([3]int,struct{X int},error){only()!;return [3]int{7},struct{X int}{8},nil}

 type Same union { First int; Second int }
 type GenericUnion[T any] union { Value T; Empty struct{} }
 func unionRead(u U)int{switch p:=u.(variant){case A:return p;case B:return len(p)}}
 func unionSame(u Same)int{switch p:=u.(variant){case First:return p;case Second:return p+1}}
 func unionDefault(u U)int{switch p:=u.(variant){case A:return p;default:_=p;return 9}}
 func unionUnbound(u U)int{switch u.(variant){case A:return 1;case B:return 2}}
 func unionCtorValue()func(int)U{return U.A}
 func unionGeneric[T any](x T)GenericUnion[T]{return GenericUnion[T].Value(x)}
 func unionSwitchBang()(int,error){switch p:=unionBang()!.(variant){case A:return p,nil;case B:return len(p),nil}}

 func slicePair()([]int,int,error){return []int{1},2,nil}
 func sliceCopy()([]int,[]int,error){return make([]int,1),[]int{7},nil}
 func mapKey(m map[int]int)(map[int]int,int,error){return m,7,nil}
 func twoInts()(int,int,error){return 7,8,nil}
 func sum(xs ...int)int{n:=0;for _,v:=range xs{n+=v};return n}
 func appendPair()([]int,error){return append(slicePair()!),nil}
 func copyPair()(int,error){return copy(sliceCopy()!),nil}
 func deletePair(m map[int]int)error{delete(mapKey(m)!);return nil}
 func variadic()(int,error){return sum(twoInts()!),nil}
 func main(){
  sl,se:=appendPair();if se!=nil||len(sl)!=2||sl[1]!=2{panic("append tuple")}
  copied,ce:=copyPair();if copied!=1||ce!=nil{panic("copy tuple")}
  m:=map[int]int{7:8};de:=deletePair(m);if de!=nil||len(m)!=0{panic("delete tuple")}
  added,ve:=variadic();if added!=15||ve!=nil{panic("variadic tuple")}


  g,ge:=generic(9,true);if g!=0||ge!=boom{panic("generic")}
  a,st,ae:=arrays();if a!=[3]int{}||st.X!=0||ae!=boom{panic("arrays")}

  n,e:=unnamed(true);if n!=0||e!=boom{panic("unnamed")}
  n,e=unnamed(false);if n!=7||e!=nil{panic("success")}
  n,e=named();if n!=43||e!=boom{panic("named")}
  n,e=tuple();if n!=9||e!=nil{panic("tuple argument")}
  n,e=iterator();if n!=42||e!=boom{panic("iterator")}
  if discard()!=boom{panic("discard")}
  n,e=tail();if n!=7||e!=nil{panic("tail tuple")}
 }
`

func parse() (*token.FileSet, []*ast.File) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "bang.go", source, 0)
	if err != nil {
		panic(err)
	}
	return fs, []*ast.File{f}
}

func main() {
	fs, files := parse()
	sp, _, err := ssautil.BuildPackage(new(types.Config), fs, types.NewPackage("main", "main"), files, ssa.SanityCheckFunctions|ssa.InstantiateGenerics)
	if err != nil {
		panic(err)
	}
	fs, files = parse()
	ip, _, err := irutil.BuildPackage(new(types.Config), fs, types.NewPackage("main", "main"), files, ir.SanityCheckFunctions)
	if err != nil {
		panic(err)
	}
	for _, name := range []string{"unnamed", "named", "tuple", "iterator", "discard", "tail"} {
		var sb, ib bytes.Buffer
		var ssaDump func(*ssa.Function)
		ssaDump = func(fn *ssa.Function) {
			fn.WriteTo(&sb)
			for _, child := range fn.AnonFuncs {
				ssaDump(child)
			}
		}
		var irDump func(*ir.Function)
		irDump = func(fn *ir.Function) {
			fn.WriteTo(&ib)
			for _, child := range fn.AnonFuncs {
				irDump(child)
			}
		}
		ssaDump(sp.Func(name))
		irDump(ip.Func(name))
		for engine, dump := range map[string]string{"ssa": sb.String(), "staticcheck": ib.String()} {
			if !strings.Contains(dump, "if ") {
				panic(fmt.Sprintf("%s %s omits error branch:\n%s", engine, name, dump))
			}
		}
	}
	if code := interp.Interpret(sp, 0, types.SizesFor("gc", "amd64"), "bang.go", nil); code != 0 {
		panic(fmt.Sprintf("SSA execution failed: %d", code))
	}
	fmt.Println("gopls SSA and Staticcheck propagation tests passed")
}
