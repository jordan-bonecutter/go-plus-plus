// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"
	"reflect"
	"runtime"
	"unsafe"
)

type Choice[T any] union {
	Some T
	None struct{}
}
type Pair[A, B any] union {
	First  A
	Second B
}
type Alias[T any] = Choice[T]
type union int

func ordinaryUnion() union { return 42 }

type Node union {
	Link *Node
	End  int
}

func assert(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func panics(f func()) (result bool)  { defer func() { result = recover() != nil }(); f(); return }
func same[T comparable](a, b T) bool { return a == b }

func (c *Choice[T]) Set(v T) { *c = Choice[T].Some(v) }
func (c Choice[T]) Value() T {
	switch v := c.(variant) {
	case Some:
		return v
	case None:
		var z T
		return z
	}
}

func testGenerics() {
	assert([]union{ordinaryUnion()}[0] == 42, "ordinary union identifier")
	var v Alias[int]
	var setter interface{ Set(int) } = &v
	setter.Set(17)
	assert(v.Value() == 17 && same(v, Choice[int].Some(17)), "generic alias and pointer method")
	constructor := Pair[int, string].Second
	switch v := constructor("text").(variant) {
	case First:
		panic(v)
	case Second:
		assert(v == "text", "two type parameters")
	}
	nested := Choice[Choice[string]].Some(Choice[string].Some("nested"))
	assert(nested.Value().Value() == "nested", "nested generic payload")
	assert(same(nested, Choice[Choice[string]].Some(Choice[string].Some("nested"))), "nested generic equality")
	type Blob [257]byte
	var b Blob
	b[256] = 99
	large := Choice[Blob].Some(b)
	assert(large.Value() == b, "large generic copy")
	assert(unsafe.Sizeof(large) == (uintptr(len(b))+unsafe.Sizeof(uintptr(0))-1)&^(unsafe.Sizeof(uintptr(0))-1)+2*unsafe.Sizeof(uintptr(0)), "large alignment")
}

func testControlFlow() {
	v := Choice[int].Some(9)
	count := 0
	switch init := 3; data := v.(variant) {
	case Some:
		assert(init+data == 12, "switch init scope")
	default:
		var _ Choice[int] = data
		panic("unexpected default")
	}
outer:
	for i := 0; i < 3; i++ {
		switch data := v.(variant) {
		case Some:
			count += data
			if i == 0 {
				continue outer
			}
			break
		case None:
			panic(data)
		}
		count++
	}
	assert(count == 29, "break and labeled continue")
match:
	switch data := v.(variant) {
	case Some:
		if data == 9 {
			break match
		}
		panic("label")
	case None:
		panic(data)
	}
	switch data := Choice[int].None(struct{}{}).(variant) {
	default:
		assert(!same(data, v), "default union binding")
	case Some:
		panic(data)
	}
	switch v.(variant) {
	case Some:
	case None:
	}
	type variant = int
	var ordinary any = 42
	assert(ordinary.(variant) == 42, "contextual variant identifier")
}

func testMaps() {
	keys := make(map[Choice[int]]int)
	for i := 0; i < 2048; i++ {
		keys[Choice[int].Some(i)] = i + 1
	}
	keys[Choice[int].None(struct{}{})] = -1
	runtime.GC()
	for i := 0; i < 2048; i++ {
		assert(keys[Choice[int].Some(i)] == i+1, "map growth lookup")
		delete(keys, Choice[int].Some(i))
	}
	assert(len(keys) == 1, "map deletion")
	type Float union {
		Number float64
		Other  float64
	}
	negative := math.Copysign(0, -1)
	floats := map[Float]int{Float.Number(negative): 1}
	floats[Float.Number(0)] = 2
	assert(len(floats) == 1 && floats[Float.Number(negative)] == 2, "signed zero map keys")
	nan := Float.Number(math.NaN())
	assert(nan != nan, "NaN is nonreflexive")
	floats[nan] = 3
	floats[nan] = 4
	assert(len(floats) == 3 && floats[nan] == 0, "NaN map keys")
	type Array [40]Choice[string]
	var a, b Array
	for i := range a {
		a[i] = Choice[string].Some("same")
		b[i] = Choice[string].Some(string([]byte("same")))
	}
	assert(a == b, "equality signature array loop")
	m := map[Array]int{a: 5}
	assert(m[b] == 5, "hash signature array loop")
	b[39] = Choice[string].None(struct{}{})
	assert(a != b && m[b] == 0, "last array variant")
	// Runtime-created maps use runtime typehash instead of the static hasher.
	typ := reflect.MapOf(reflect.TypeOf(Choice[string]{}), reflect.TypeOf(complex64(0)))
	dynamic := reflect.MakeMap(typ)
	dynamic.SetMapIndex(reflect.ValueOf(Choice[string]{}), reflect.ValueOf(complex64(7)))
	got := dynamic.MapIndex(reflect.ValueOf(Choice[string].Some("")))
	assert(got.IsValid() && got.Complex() == 7, "reflect map zero key")
	boxed := map[any]int{Choice[string].Some("boxed"): 8}
	assert(boxed[Choice[string].Some(string([]byte("boxed")))] == 8, "boxed map key")
}

func testComparisonOrder() {
	bad := Choice[any].Some([]int{1})
	type First struct {
		Value Choice[any]
		Tail  int
	}
	assert(panics(func() { a, b := First{bad, 1}, First{bad, 2}; _ = a == b }), "union comparison panic order")
	type Last struct {
		Head  int
		Value Choice[any]
	}
	assert(!panics(func() { a, b := Last{1, bad}, Last{2, bad}; _ = a == b }), "union comparison short circuit")
}

func testEmptyInterfaceEquality() {
	type Empty union {
		A struct{}
		B [0]int
	}
	a, b := Empty.A(struct{}{}), Empty.A(struct{}{})
	assert(any(a) == any(b) && any(Empty{}) == any(a), "empty boxed equality")
	assert(any(a) != any(Empty.B([0]int{})), "empty boxed variant identity")
	type Wrapped struct {
		Before [0]string
		Value  [1]Empty
	}
	x, y := Wrapped{Value: [1]Empty{a}}, Wrapped{Value: [1]Empty{b}}
	assert(any(x) == any(y), "wrapped empty boxed equality")
	assert(map[any]int{a: 1}[b] == 1 && map[any]int{x: 2}[y] == 2, "empty boxed map keys")
	// These types are constructed at runtime, so reflection must propagate the
	// union's indirect interface representation through wrappers as well.
	for _, typ := range []reflect.Type{
		reflect.StructOf([]reflect.StructField{{Name: "Payload", Type: reflect.TypeOf(a)}}),
		reflect.ArrayOf(1, reflect.TypeOf(a)),
	} {
		left, right := reflect.New(typ).Elem(), reflect.New(typ).Elem()
		if typ.Kind() == reflect.Struct {
			left.Field(0).Set(reflect.ValueOf(a))
			right.Field(0).Set(reflect.ValueOf(b))
		} else {
			left.Index(0).Set(reflect.ValueOf(a))
			right.Index(0).Set(reflect.ValueOf(b))
		}
		assert(left.Interface() == right.Interface(), "runtime wrapper boxed equality")
		assert(map[any]int{left.Interface(): 1}[right.Interface()] == 1, "runtime wrapper map keys")
	}

}

func testReflection() {
	zero := Choice[string]{}
	explicit := Choice[string].Some("")
	assert(reflect.DeepEqual(zero, explicit), "DeepEqual zero constructor")
	type Slices union {
		Data  []int
		Empty struct{}
	}
	assert(reflect.DeepEqual(Slices.Data([]int{1, 2}), Slices.Data([]int{1, 2})), "DeepEqual slice payload")
	assert(!reflect.DeepEqual(Slices.Data([]int{1}), Slices.Data([]int{2})), "DeepEqual different slice payload")
	type Repeated union {
		A int
		B int
	}
	assert(!reflect.DeepEqual(Repeated.A(1), Repeated.B(1)), "DeepEqual variant identity")
	type Empty union {
		First  struct{}
		Second struct{}
	}
	assert(reflect.ValueOf(Empty{}).Equal(reflect.ValueOf(Empty.First(struct{}{}))), "direct interface zero union")
	assert(!reflect.ValueOf(Empty{}).Equal(reflect.ValueOf(Empty.Second(struct{}{}))), "direct interface variant identity")
	type Dynamic union {
		Value any
		Empty struct{}
	}
	var dynamicZero Dynamic
	assert(dynamicZero == Dynamic.Value(nil), "nil interface payload")
	assert(Dynamic.Value((*int)(nil)) != dynamicZero, "typed nil interface payload")
	assert(Dynamic.Value(0) != Dynamic.Value(int64(0)), "interface dynamic type")
	assert(map[Dynamic]int{dynamicZero: 1}[Dynamic.Value(nil)] == 1, "nil interface map key")
	bad := Dynamic.Value([]int{1})
	assert(bad != Dynamic.Empty(struct{}{}), "different variants avoid interface panic")
	assert(!reflect.ValueOf(bad).Comparable(), "dynamic Comparable")
	assert(panics(func() { reflect.ValueOf(bad).Equal(reflect.ValueOf(bad)) }), "dynamic Equal panic")
	assert(!reflect.TypeOf(Slices{}).Comparable(), "noncomparable union type")
	assert(!reflect.ValueOf([0]Slices{}).Comparable(), "zero array of noncomparable union")
	assert(panics(func() { v := reflect.ValueOf([0]Slices{}); v.Equal(v) }), "zero array union comparison panic")
	assert(reflect.ValueOf(Slices.Data(nil)).IsZero(), "noncomparable union IsZero")
	assert(!reflect.ValueOf(Slices.Data([]int{})).IsZero(), "nonnull slice IsZero")
	assert(!reflect.ValueOf(Slices.Empty(struct{}{})).IsZero(), "nonfirst empty variant IsZero")
	type LargeZero union {
		Data  [2048]int
		Empty struct{}
	}
	assert(reflect.ValueOf(LargeZero.Data([2048]int{})).IsZero(), "large union IsZero")
	assert(!reflect.ValueOf(LargeZero.Empty(struct{}{})).IsZero(), "large variant IsZero")
	assert(panics(func() { reflect.ValueOf(Slices.Empty(struct{}{})).Equal(reflect.ValueOf(Slices.Empty(struct{}{}))) }), "inactive noncomparable variant")
}

//go:noinline
func integer(n int) *int { p := new(int); *p = n; return p }

//go:noinline
func closure(n int) func() int { p := integer(n); return func() int { return *p } }

//go:noinline
func denseRoots() Choice[[257]*int] {
	var roots [257]*int
	for i := range roots {
		roots[i] = integer(i + 1)
	}
	return Choice[[257]*int].Some(roots)
}

type SparseRoots struct {
	Padding [256]uintptr
	Pointer *int
}

//go:noinline
func sparseRoots() Choice[SparseRoots] {
	return Choice[SparseRoots].Some(SparseRoots{Pointer: integer(99)})
}

func testDemandGCMasks() {
	// Both payloads exceed the runtime's inline bitmap limit. One has dense
	// pointer words; the other has a single pointer beyond a large scalar prefix.
	dense, sparse := denseRoots(), sparseRoots()
	for i := 0; i < 4; i++ {
		runtime.GC()
	}
	for i, p := range dense.Value() {
		assert(*p == i+1, "dense GC mask")
	}
	assert(*sparse.Value().Pointer == 99, "sparse GC mask")
}

func testGCAndCopies() {
	type Resources union {
		Slice    []*int
		Map      map[string]*int
		Function func() int
		Channel  chan *int
	}
	ch := make(chan *int, 1)
	ch <- integer(74)
	values := []Resources{Resources.Slice([]*int{integer(71)}), Resources.Map(map[string]*int{"x": integer(72)}), Resources.Function(closure(73)), Resources.Channel(ch)}
	runtime.GC()
	for _, value := range values {
		switch data := value.(variant) {
		case Slice:
			assert(*data[0] == 71, "slice roots")
		case Map:
			assert(*data["x"] == 72, "map roots")
		case Function:
			assert(data() == 73, "closure roots")
		case Channel:
			assert(*<-data == 74, "channel roots")
		}
	}
	left, right := Node{}, Node{}
	left, right = Node.Link(&left), Node.Link(&right)
	assert(reflect.DeepEqual(left, right), "cyclic union DeepEqual")
	root := Node.End(7)
	for i := 0; i < 64; i++ {
		previous := root
		root = Node.Link(&previous)
	}
	runtime.GC()
	for i := 0; i < 64; i++ {
		switch data := root.(variant) {
		case Link:
			root = *data
		case End:
			panic("short recursive chain")
		}
	}
	switch data := root.(variant) {
	case End:
		assert(data == 7, "recursive payload")
	case Link:
		panic("long recursive chain")
	}
	// Extracted payloads must survive after their union anchor becomes unreachable.
	extracted := func() []*int {
		switch data := values[0].(variant) {
		case Slice:
			return data
		default:
			panic("slice")
		}
	}()
	values = nil
	runtime.GC()
	assert(*extracted[0] == 71, "extracted roots")
}

func testConcurrentCopies() {
	values := make(chan Choice[*[64]byte], 3)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 5; i++ {
			runtime.GC()
		}
	}()
	go func() {
		defer close(values)
		for i := 0; i < 200; i++ {
			p := new([64]byte)
			p[0], p[63] = byte(i), byte(i+1)
			values <- Choice[*[64]byte].Some(p)
		}
	}()
	i := 0
	for value := range values {
		p := value.Value()
		assert(p[0] == byte(i) && p[63] == byte(i+1), "concurrent channel copy")
		i++
	}
	assert(i == 200, "channel count")
	<-done
}

func main() {
	testDemandGCMasks()
	testComparisonOrder()
	testEmptyInterfaceEquality()
	testConcurrentCopies()
	testGenerics()
	testControlFlow()
	testMaps()
	testReflection()
	testGCAndCopies()
}
