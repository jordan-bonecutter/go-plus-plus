// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xml_test

import (
	"encoding/xml"
	"reflect"
	"testing"
)

type unionChoice union {
	Number  int
	Again   int
	Text    string
	List    []unionChoice
	Pointer *int
	hidden  int
}

type UnionEmbedded union{ Value int }

func TestUnionXML(t *testing.T) {
	n := 42
	for _, tt := range []struct {
		value unionChoice
		want  string
	}{
		{unionChoice{}, `<unionChoice><Number><value>0</value></Number></unionChoice>`},
		{unionChoice.Again(0), `<unionChoice><Again><value>0</value></Again></unionChoice>`},
		{unionChoice.Text("hello"), `<unionChoice><Text><value>hello</value></Text></unionChoice>`},
		{unionChoice.List(nil), `<unionChoice><List></List></unionChoice>`},
		{unionChoice.Pointer(nil), `<unionChoice><Pointer></Pointer></unionChoice>`},
		{unionChoice.Pointer(&n), `<unionChoice><Pointer><value>42</value></Pointer></unionChoice>`},
		{unionChoice.List([]unionChoice{unionChoice.Again(3)}), `<unionChoice><List><value><Again><value>3</value></Again></value></List></unionChoice>`},
	} {
		data, err := xml.Marshal(tt.value)
		if err != nil || string(data) != tt.want {
			t.Fatalf("Marshal: %s %v; want %s", data, err, tt.want)
		}
		got := unionChoice.Text("old")
		if err := xml.Unmarshal(data, &got); err != nil || !reflect.DeepEqual(got, tt.value) {
			t.Fatalf("round trip %v: %v %v", tt.value, got, err)
		}
	}
	for _, input := range []string{`<x/>`, `<x><Missing/></x>`, `<x><hidden/></x>`, `<x><Number/><Again/></x>`, `<x><Number><value>bad</value></Number></x>`, `<x>text<Number/></x>`} {
		before := unionChoice.Text("old")
		got := before
		if err := xml.Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("accepted %s", input)
		}
		if !reflect.DeepEqual(got, before) {
			t.Errorf("mutated destination for %s", input)
		}
	}
	if _, err := xml.Marshal(unionChoice.hidden(1)); err == nil {
		t.Error("marshaled hidden variant")
	}
	wrapped := struct {
		XMLName xml.Name
		UnionEmbedded
	}{XMLName: xml.Name{Local: "root"}, UnionEmbedded: UnionEmbedded.Value(7)}
	data, err := xml.Marshal(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		XMLName xml.Name
		UnionEmbedded
	}
	if err := xml.Unmarshal(data, &out); err != nil || out != wrapped {
		t.Fatalf("embedded: %s %v %v", data, out, err)
	}
}

func TestUnionXMLNamespace(t *testing.T) {
	want := struct {
		Choice unionChoice `xml:"urn:test choice"`
	}{unionChoice.Again(7)}
	data, err := xml.Marshal(struct {
		XMLName xml.Name
		Choice  unionChoice `xml:"urn:test choice"`
	}{xml.Name{Local: "root"}, want.Choice})
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Choice unionChoice `xml:"urn:test choice"`
	}
	if err := xml.Unmarshal(data, &got); err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("namespace round trip: %s %v %v", data, got, err)
	}
}
