// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gob_test

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"
)

type unionChoice union {
	Number  int
	Again   int
	Text    string
	List    []unionChoice
	Map     map[string]int
	Any     any
	Pointer *int
	hidden  int
}

type unionCustom union{ Value int }

func (unionCustom) GobEncode() ([]byte, error) { return []byte{7}, nil }
func (u *unionCustom) GobDecode([]byte) error  { *u = unionCustom.Value(7); return nil }

func TestUnionGob(t *testing.T) {
	n := 42
	for _, want := range []unionChoice{unionChoice{}, unionChoice.Again(0), unionChoice.Text("hello"), unionChoice.List([]unionChoice{unionChoice.Again(4)}), unionChoice.Map(map[string]int{"x": 3}), unionChoice.Any(nil), unionChoice.Any(4), unionChoice.Pointer(&n)} {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		dec := gob.NewDecoder(&buf)
		// Reuse the engines and replace an already initialized destination.
		got := unionChoice.Text("old")
		for range 2 {
			if err := enc.Encode(want); err != nil {
				t.Fatalf("Encode(%v): %v", want, err)
			}
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("Decode(%v): %v", want, err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
		buf.Reset()
		wrapped := struct {
			Items []unionChoice
			Key   map[string]unionChoice
		}{[]unionChoice{want}, map[string]unionChoice{"x": want}}
		if err := gob.NewEncoder(&buf).Encode(wrapped); err != nil {
			t.Fatal(err)
		}
		var out struct {
			Items []unionChoice
			Key   map[string]unionChoice
		}
		if err := gob.NewDecoder(&buf).Decode(&out); err != nil || !reflect.DeepEqual(out, wrapped) {
			t.Fatalf("nested round trip: %v %v", out, err)
		}
	}
	for _, bad := range []unionChoice{unionChoice.hidden(1), unionChoice.Pointer(nil)} {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(bad); err == nil {
			t.Errorf("accepted %v", bad)
		}
	}
	for _, wire := range []any{struct {
		Number int
		Text   string
	}{1, "two"}, struct{ Missing int }{1}, struct{ Number int }{0}} {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(wire); err != nil {
			t.Fatal(err)
		}
		before := unionChoice.Text("unchanged")
		got := before
		if err := gob.NewDecoder(&buf).Decode(&got); err == nil {
			t.Errorf("accepted invalid union %v", wire)
		}
		if !reflect.DeepEqual(got, before) {
			t.Errorf("mutated destination: %v", got)
		}
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(unionCustom{}); err != nil {
		t.Fatal(err)
	}
	var custom unionCustom
	if err := gob.NewDecoder(&buf).Decode(&custom); err != nil || custom != unionCustom.Value(7) {
		t.Fatalf("custom codec: %v %v", custom, err)
	}
}
