// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"runtime"
)

type Choice[T any] union {
	First  T
	Second string
	Again  T
	hidden int
}

type Dynamic union {
	Any   any
	Bytes []byte
}

func check(ok bool) {
	if !ok {
		panic("union reflection")
	}
}
func panics(f func()) { defer func() { check(recover() != nil) }(); f() }

func main() {
	testWideTag()
	typ := reflect.TypeFor[Choice[*int]]()
	check(typ.IsUnion() && typ.NumVariant() == 4)
	check(!reflect.TypeFor[int]().IsUnion() && !reflect.TypeFor[struct{}]().IsUnion())
	for i, name := range []string{"First", "Second", "Again", "hidden"} {
		variant := typ.Variant(i)
		byName, ok := typ.VariantByName(name)
		check(ok && variant == byName && variant.Name == name && variant.Index == i)
		check(variant.IsExported() == (i != 3))
		if i == 0 || i == 2 {
			check(variant.Type == reflect.TypeFor[*int]())
		}
	}
	_, ok := typ.VariantByName("missing")
	check(!ok)
	panics(func() { typ.Variant(-1) })
	panics(func() { typ.Variant(4) })
	panics(func() { reflect.TypeFor[int]().NumVariant() })
	panics(func() { reflect.TypeFor[int]().VariantByName("x") })
	value := reflect.New(typ).Elem()
	check(value.IsZero())
	n := new(int)
	*n = 42
	value.SetVariant("Again", reflect.ValueOf(n))
	name, payload := value.Variant()
	check(name == "Again" && !payload.CanSet() && !payload.CanAddr() && payload.Interface().(*int) == n)
	check(value.Interface() == Choice[*int].Again(n))
	// Returned payloads are independent snapshots with their own GC roots.
	value.SetVariant("Second", reflect.ValueOf("changed"))
	n = nil
	runtime.GC()
	check(*payload.Interface().(*int) == 42)
	check(!value.IsZero())
	before := value.Interface()
	panics(func() { value.SetVariant("First", reflect.ValueOf(1)) })
	panics(func() { value.SetVariant("missing", reflect.ValueOf(1)) })
	panics(func() { value.SetVariant("hidden", reflect.ValueOf(1)) })
	panics(func() { reflect.ValueOf(before).SetVariant("Second", reflect.ValueOf("x")) })
	check(value.Interface() == before)
	value.SetZero()
	check(value.IsZero())
	value.SetVariant("First", reflect.Zero(reflect.TypeFor[*int]()))
	check(value.IsZero() && value.Equal(reflect.Zero(typ)))
	dyn := reflect.New(reflect.TypeFor[Dynamic]()).Elem()
	dyn.SetVariant("Any", reflect.ValueOf([]int{1, 2}))
	check(!dyn.Comparable())
	_, data := dyn.Variant()
	check(reflect.DeepEqual(data.Interface(), []int{1, 2}))
	dyn.SetVariant("Bytes", reflect.ValueOf([]byte{3, 4}))
	runtime.GC()
	check(reflect.DeepEqual(dyn.Interface(), Dynamic.Bytes([]byte{3, 4})))
	// Exercise ordinary reflective calls, containers, copies, and conversions.
	fn := reflect.ValueOf(func(x Choice[*int]) Choice[*int] { return x })
	result := fn.Call([]reflect.Value{value})[0]
	check(result.Equal(value))
	m := reflect.MakeMap(reflect.MapOf(typ, typ))
	m.SetMapIndex(value, result)
	check(m.MapIndex(value).Equal(value))
	slice := reflect.MakeSlice(reflect.SliceOf(typ), 1, 1)
	slice.Index(0).Set(value)
	check(slice.Index(0).Equal(value))
	check(value.Convert(typ).Equal(value))
}

// The 257th variant crosses the uint8 boundary.
type Wide union {
	V0   int
	V1   int
	V2   int
	V3   int
	V4   int
	V5   int
	V6   int
	V7   int
	V8   int
	V9   int
	V10  int
	V11  int
	V12  int
	V13  int
	V14  int
	V15  int
	V16  int
	V17  int
	V18  int
	V19  int
	V20  int
	V21  int
	V22  int
	V23  int
	V24  int
	V25  int
	V26  int
	V27  int
	V28  int
	V29  int
	V30  int
	V31  int
	V32  int
	V33  int
	V34  int
	V35  int
	V36  int
	V37  int
	V38  int
	V39  int
	V40  int
	V41  int
	V42  int
	V43  int
	V44  int
	V45  int
	V46  int
	V47  int
	V48  int
	V49  int
	V50  int
	V51  int
	V52  int
	V53  int
	V54  int
	V55  int
	V56  int
	V57  int
	V58  int
	V59  int
	V60  int
	V61  int
	V62  int
	V63  int
	V64  int
	V65  int
	V66  int
	V67  int
	V68  int
	V69  int
	V70  int
	V71  int
	V72  int
	V73  int
	V74  int
	V75  int
	V76  int
	V77  int
	V78  int
	V79  int
	V80  int
	V81  int
	V82  int
	V83  int
	V84  int
	V85  int
	V86  int
	V87  int
	V88  int
	V89  int
	V90  int
	V91  int
	V92  int
	V93  int
	V94  int
	V95  int
	V96  int
	V97  int
	V98  int
	V99  int
	V100 int
	V101 int
	V102 int
	V103 int
	V104 int
	V105 int
	V106 int
	V107 int
	V108 int
	V109 int
	V110 int
	V111 int
	V112 int
	V113 int
	V114 int
	V115 int
	V116 int
	V117 int
	V118 int
	V119 int
	V120 int
	V121 int
	V122 int
	V123 int
	V124 int
	V125 int
	V126 int
	V127 int
	V128 int
	V129 int
	V130 int
	V131 int
	V132 int
	V133 int
	V134 int
	V135 int
	V136 int
	V137 int
	V138 int
	V139 int
	V140 int
	V141 int
	V142 int
	V143 int
	V144 int
	V145 int
	V146 int
	V147 int
	V148 int
	V149 int
	V150 int
	V151 int
	V152 int
	V153 int
	V154 int
	V155 int
	V156 int
	V157 int
	V158 int
	V159 int
	V160 int
	V161 int
	V162 int
	V163 int
	V164 int
	V165 int
	V166 int
	V167 int
	V168 int
	V169 int
	V170 int
	V171 int
	V172 int
	V173 int
	V174 int
	V175 int
	V176 int
	V177 int
	V178 int
	V179 int
	V180 int
	V181 int
	V182 int
	V183 int
	V184 int
	V185 int
	V186 int
	V187 int
	V188 int
	V189 int
	V190 int
	V191 int
	V192 int
	V193 int
	V194 int
	V195 int
	V196 int
	V197 int
	V198 int
	V199 int
	V200 int
	V201 int
	V202 int
	V203 int
	V204 int
	V205 int
	V206 int
	V207 int
	V208 int
	V209 int
	V210 int
	V211 int
	V212 int
	V213 int
	V214 int
	V215 int
	V216 int
	V217 int
	V218 int
	V219 int
	V220 int
	V221 int
	V222 int
	V223 int
	V224 int
	V225 int
	V226 int
	V227 int
	V228 int
	V229 int
	V230 int
	V231 int
	V232 int
	V233 int
	V234 int
	V235 int
	V236 int
	V237 int
	V238 int
	V239 int
	V240 int
	V241 int
	V242 int
	V243 int
	V244 int
	V245 int
	V246 int
	V247 int
	V248 int
	V249 int
	V250 int
	V251 int
	V252 int
	V253 int
	V254 int
	V255 int
	V256 any
}

func testWideTag() {
	t := reflect.TypeFor[Wide]()
	check(t.Field(2).Type.Kind() == reflect.Uint16)
	v := Wide.V256("wide")
	switch x := v.(variant) {
	case V256:
		check(x == "wide")
	default:
		panic("wide variant dispatch")
	}
	check(v == Wide.V256("wide"))
	check(map[Wide]int{v: 42}[Wide.V256("wide")] == 42)
	r := reflect.New(t).Elem()
	r.SetVariant("V256", reflect.ValueOf("wide"))
	check(r.Interface() == v)
	name, payload := r.Variant()
	check(name == "V256" && payload.Interface() == "wide")
	check(reflect.DeepEqual(r.Interface(), v))
	r.SetVariant("V255", reflect.ValueOf(255))
	check(r.Interface() == Wide.V255(255))
	r.SetVariant("V256", reflect.ValueOf([]byte{1}))
	check(!r.Comparable())
	// Failed dynamic hashing must inspect the active payload on every map path.
	key := r.Interface().(Wide)
	panics(func() { _ = map[Wide]int{}[key] })
	panics(func() { _ = reflect.MakeMap(reflect.MapOf(t, reflect.TypeFor[int]())).MapIndex(r) })
}
