// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.jsonv2

package json_test

import (
	"encoding/json/v2"
	"reflect"
	"testing"
)

type unionChoice union {
	Number int
	Again  int
	List   []unionChoice
	hidden int
}

func TestUnion(t *testing.T) {
	for _, want := range []unionChoice{unionChoice{}, unionChoice.Again(0), unionChoice.List([]unionChoice{unionChoice.Again(4)})} {
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		var got unionChoice
		if err := json.Unmarshal(data, &got); err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("round trip %s: %v %v", data, got, err)
		}
	}
	before := unionChoice.Again(7)
	for _, input := range []string{`{}`, `{"missing":1}`, `{"hidden":1}`, `{"Number":1,"Again":2}`, `{"Number":1,"Number":2}`, `{"Number":"bad"}`} {
		got := before
		if err := json.Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("accepted %s", input)
		}
		if !reflect.DeepEqual(got, before) {
			t.Errorf("mutated destination for %s", input)
		}
	}
	marshalers := json.MarshalFunc(func(v int) ([]byte, error) { return []byte(`"custom"`), nil })
	data, err := json.Marshal(before, json.WithMarshalers(marshalers))
	if err != nil || string(data) != `{"Again":"custom"}` {
		t.Fatalf("custom payload marshal: %s %v", data, err)
	}
	unmarshalers := json.UnmarshalFunc(func(data []byte, v *int) error { *v = 42; return nil })
	var got unionChoice
	if err := json.Unmarshal(data, &got, json.WithUnmarshalers(unmarshalers)); err != nil || !reflect.DeepEqual(got, unionChoice.Again(42)) {
		t.Fatalf("custom payload unmarshal: %v %v", got, err)
	}
	got = before
	if err := json.Unmarshal([]byte(`null`), &got); err != nil || !reflect.DeepEqual(got, unionChoice{}) {
		t.Fatalf("null reset: %v %v", got, err)
	}
}

func TestUnionStringTag(t *testing.T) {
	value := struct {
		Choice unionChoice `json:",string"`
	}{unionChoice.Again(7)}
	if _, err := json.Marshal(value); err == nil {
		t.Fatal("accepted string tag on union")
	}
	if err := json.Unmarshal([]byte(`{"Choice":{"Again":7}}`), &value); err == nil {
		t.Fatal("accepted string tag on union")
	}
	data, err := json.Marshal(unionChoice.Again(7), json.StringifyNumbers(true))
	if err != nil || string(data) != `{"Again":"7"}` {
		t.Fatalf("StringifyNumbers: %s %v", data, err)
	}
}
