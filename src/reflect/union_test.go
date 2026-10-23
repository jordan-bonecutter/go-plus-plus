// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect_test

import (
	"math"
	"reflect"
	"runtime"
	"testing"
)

type kindChoice union {
	Number  int
	Text    string
	Dynamic any
}
type kindSlice union {
	Items  []int
	Number int
}

func (c kindChoice) Label() string { return "choice" }

func TestUnionKind(t *testing.T) {
	typ := reflect.TypeFor[kindChoice]()
	if typ.Kind() != reflect.Union || reflect.ValueOf(kindChoice{}).Kind() != reflect.Union || reflect.Union.String() != "union" || !typ.IsUnion() {
		t.Fatalf("union kind: %v %v", typ.Kind(), typ.IsUnion())
	}
	if reflect.TypeFor[struct{}]().Kind() != reflect.Struct || reflect.TypeFor[struct{}]().IsUnion() {
		t.Fatal("struct mistaken for union")
	}
	if typ.NumVariant() != 3 || typ.Variant(1).Name != "Text" {
		t.Fatal("variant metadata")
	}
	for name, f := range map[string]func(){
		"Type.NumField":    func() { typ.NumField() },
		"Type.Field":       func() { typ.Field(0) },
		"Type.FieldByName": func() { typ.FieldByName("Number") },
		"Value.NumField":   func() { reflect.ValueOf(kindChoice{}).NumField() },
		"Value.Field":      func() { reflect.ValueOf(kindChoice{}).Field(0) },
		"VisibleFields":    func() { reflect.VisibleFields(typ) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			f()
		})
	}
	if name, _ := reflect.ValueOf(struct{}{}).Variant(); name != "" {
		t.Fatal("ordinary struct has variant")
	}
}

func TestUnionKindValues(t *testing.T) {
	v := reflect.New(reflect.TypeFor[kindChoice]()).Elem()
	if !v.IsZero() {
		t.Fatal("zero")
	}
	v.SetVariant("Text", reflect.ValueOf("hello"))
	runtime.GC()
	if v.IsZero() || v.Interface() != kindChoice.Text("hello") {
		t.Fatal("set variant")
	}
	n, p := v.Variant()
	if n != "Text" || p.String() != "hello" {
		t.Fatal(n, p)
	}
	if !v.Comparable() || !v.Equal(reflect.ValueOf(kindChoice.Text("hello"))) {
		t.Fatal("equal")
	}
	v.SetZero()
	if !v.IsZero() || v.Interface() != kindChoice.Number(0) {
		t.Fatal("set zero")
	}
	a, b := kindSlice.Items([]int{1, 2}), kindSlice.Items([]int{1, 2})
	if !reflect.DeepEqual(a, b) || reflect.DeepEqual(a, kindSlice.Number(1)) || reflect.ValueOf(a).Comparable() {
		t.Fatal("deep equal/comparable")
	}
	bad := kindChoice.Dynamic([]int{1})
	if reflect.ValueOf([1]kindChoice{bad}).Comparable() {
		t.Fatal("array of dynamically noncomparable union")
	}
}

func TestUnionKindCalls(t *testing.T) {
	x := kindChoice.Text("hello")
	fn := reflect.ValueOf(func(c kindChoice) kindChoice { return c })
	if got := fn.Call([]reflect.Value{reflect.ValueOf(x)})[0].Interface(); got != x {
		t.Fatal(got)
	}
	made := reflect.MakeFunc(fn.Type(), func(v []reflect.Value) []reflect.Value { return v })
	if got := made.Interface().(func(kindChoice) kindChoice)(x); got != x {
		t.Fatal(got)
	}
	method := reflect.ValueOf(x).MethodByName("Label")
	if got := method.Call(nil)[0].String(); got != "choice" {
		t.Fatal(got)
	}
}

func TestUnionKindDynamicContainers(t *testing.T) {
	typ := reflect.TypeFor[kindChoice]()
	for _, keyType := range []reflect.Type{typ, reflect.ArrayOf(1, typ), reflect.StructOf([]reflect.StructField{{Name: "Value", Type: typ}})} {
		key := reflect.New(keyType).Elem()
		payload := key
		if keyType.Kind() == reflect.Array {
			payload = key.Index(0)
		} else if keyType.Kind() == reflect.Struct {
			payload = key.Field(0)
		}
		payload.Set(reflect.ValueOf(kindChoice.Text("hello")))
		m := reflect.MakeMap(reflect.MapOf(keyType, reflect.TypeFor[int]()))
		m.SetMapIndex(key, reflect.ValueOf(42))
		runtime.GC()
		if got := m.MapIndex(key).Int(); got != 42 {
			t.Fatal(got)
		}
		// The dynamic hash must inspect active payloads, including empty maps.
		payload.Set(reflect.ValueOf(kindChoice.Dynamic([]int{1})))
		empty := reflect.MakeMap(m.Type())
		func() {
			defer func() {
				if recover() == nil {
					t.Error("unhashable lookup did not panic")
				}
			}()
			empty.MapIndex(key)
		}()
	}
	type floating union {
		Value float64
		Text  string
	}
	ft := reflect.TypeFor[floating]()
	m := reflect.MakeMap(reflect.MapOf(ft, reflect.TypeFor[int]()))
	nan := reflect.ValueOf(floating.Value(math.NaN()))
	for i := 0; i < 100; i++ {
		m.SetMapIndex(nan, reflect.ValueOf(i))
	}
	if m.Len() != 100 {
		t.Fatal("NaN reflexivity", m.Len())
	}
	iter := m.MapRange()
	n := 0
	for iter.Next() {
		n++
	}
	if n != 100 {
		t.Fatal("NaN iteration", n)
	}
	s := reflect.MakeSlice(reflect.SliceOf(typ), 2, 2)
	s.Index(0).Set(reflect.ValueOf(kindChoice.Text("a")))
	reflect.Copy(s.Slice(1, 2), s.Slice(0, 1))
	if s.Index(1).Interface() != kindChoice.Text("a") {
		t.Fatal("copy")
	}
	ch := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, typ), 1)
	ch.Send(s.Index(1))
	got, _ := ch.Recv()
	if !got.Equal(s.Index(0)) {
		t.Fatal("channel")
	}
}

func TestUnionKindConversionsAndKeyUpdates(t *testing.T) {
	type copyChoice kindChoice
	x := kindChoice.Text("hello")
	v := reflect.ValueOf(x)
	dst := reflect.TypeFor[copyChoice]()
	if !v.Type().ConvertibleTo(dst) || !v.CanConvert(dst) {
		t.Fatal("union conversion rejected")
	}
	if got := v.Convert(dst).Interface().(copyChoice); kindChoice(got) != x {
		t.Fatal("union conversion value")
	}
	type floating union {
		Value float64
		Text  string
	}
	m := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[floating](), reflect.TypeFor[int]()))
	m.SetMapIndex(reflect.ValueOf(floating.Value(0)), reflect.ValueOf(1))
	m.SetMapIndex(reflect.ValueOf(floating.Value(math.Copysign(0, -1))), reflect.ValueOf(2))
	if m.Len() != 1 {
		t.Fatal("signed zero equality")
	}
	_, p := m.MapKeys()[0].Variant()
	if !math.Signbit(p.Float()) {
		t.Fatal("map did not update union key")
	}
}
