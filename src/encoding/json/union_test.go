// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type unionChoice union {
	Number  int
	Text    string
	Again   int
	Any     any
	List    []unionChoice
	Pointer *int
	hidden  int
}

type UnionEmbedded union{ Value int }

type unionCustom union{ Value int }

func (unionCustom) MarshalJSON() ([]byte, error)  { return []byte(`"custom"`), nil }
func (u *unionCustom) UnmarshalJSON([]byte) error { *u = unionCustom.Value(9); return nil }

func TestUnionJSON(t *testing.T) {
	for _, tt := range []struct {
		value unionChoice
		want  string
	}{
		{unionChoice{}, `{"Number":0}`},
		{unionChoice.Again(0), `{"Again":0}`},
		{unionChoice.Text("<x>"), `{"Text":"\u003cx\u003e"}`},
		{unionChoice.Any(nil), `{"Any":null}`},
		{unionChoice.Pointer(nil), `{"Pointer":null}`},
		{unionChoice.List([]unionChoice{unionChoice.Again(3)}), `{"List":[{"Again":3}]}`},
	} {
		data, err := json.Marshal(tt.value)
		if err != nil || string(data) != tt.want {
			t.Fatalf("Marshal: %s, %v; want %s", data, err, tt.want)
		}
		var got unionChoice
		if err := json.Unmarshal(data, &got); err != nil || !reflect.DeepEqual(got, tt.value) {
			t.Fatalf("round trip: %v, %v", got, err)
		}
	}
	before := unionChoice.Text("unchanged")
	for _, input := range []string{`{}`, `{"missing":1}`, `{"hidden":1}`, `{"Number":1,"Again":2}`, `{"Number":1,"Number":2}`, `{"Number":"bad"}`, `[]`, `42`, `"Number"`} {
		got := before
		if err := json.Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("accepted %s", input)
		}
		if !reflect.DeepEqual(got, before) {
			t.Errorf("mutated destination for %s: %v", input, got)
		}
	}
	if _, err := json.Marshal(unionChoice.hidden(1)); err == nil {
		t.Error("marshaled hidden variant")
	}
	var custom unionCustom
	data, err := json.Marshal(custom)
	if err != nil || string(data) != `"custom"` {
		t.Fatalf("custom marshal: %s %v", data, err)
	}
	if err := json.Unmarshal(data, &custom); err != nil || custom != unionCustom.Value(9) {
		t.Fatalf("custom unmarshal: %v %v", custom, err)
	}
	wrapped := struct{ UnionEmbedded }{UnionEmbedded.Value(7)}
	data, err = json.Marshal(wrapped)
	if err != nil || string(data) != `{"UnionEmbedded":{"Value":7}}` {
		t.Fatalf("embedded union: %s %v", data, err)
	}
	var decoded struct{ UnionEmbedded }
	if err := json.Unmarshal(data, &decoded); err != nil || decoded != wrapped {
		t.Fatalf("embedded round trip: %v %v", decoded, err)
	}
	dec := json.NewDecoder(strings.NewReader(`{"Any":123}`))
	dec.UseNumber()
	var got unionChoice
	if err := dec.Decode(&got); err != nil {
		t.Fatal(err)
	}
	_, payload := reflect.ValueOf(got).Variant()
	if _, ok := payload.Interface().(json.Number); !ok {
		t.Fatalf("lost UseNumber: %T", payload.Interface())
	}
	dec = json.NewDecoder(bytes.NewBufferString(`{"Any":{"bad":true}}`))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&got); err != nil {
		t.Fatal(err)
	}
}

func TestUnionJSONErrorRecovery(t *testing.T) {
	for _, input := range []string{`{}`, `{"missing":1}`, `{"hidden":1}`, `{"Number":1,"Again":2}`, `{"Number":"bad"}`} {
		var got struct {
			Choice unionChoice
			Next   int
		}
		got.Choice = unionChoice.Text("old")
		if err := json.Unmarshal([]byte(`{"Choice":`+input+`,"Next":42}`), &got); err == nil {
			t.Errorf("accepted %s", input)
		}
		if got.Next != 42 || !reflect.DeepEqual(got.Choice, unionChoice.Text("old")) {
			t.Errorf("failed recovery for %s: %v", input, got)
		}
	}
}

func TestUnionJSONCycles(t *testing.T) {
	value := new(unionChoice)
	*value = unionChoice.Any(value)
	if _, err := json.Marshal(value); err == nil {
		t.Fatal("accepted cyclic union")
	}
}

func TestUnionJSONStringTag(t *testing.T) {
	want := struct {
		Choice unionChoice `json:",string"`
	}{unionChoice.Again(7)}
	data, err := json.Marshal(want)
	if err != nil || string(data) != `{"Choice":{"Again":7}}` {
		t.Fatalf("string tag: %s %v", data, err)
	}
	var got struct {
		Choice unionChoice `json:",string"`
	}
	if err := json.Unmarshal(data, &got); err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("string tag round trip: %v %v", got, err)
	}
}
