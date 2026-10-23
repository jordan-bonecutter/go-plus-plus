// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"testing"
)

type unitSignal union {
	Start
	Stop
}

func TestUnitUnionWireFormats(t *testing.T) {
	want := unitSignal.Stop()
	for _, format := range []string{"json", "gob", "xml"} {
		t.Run(format, func(t *testing.T) {
			var got unitSignal
			switch format {
			case "json":
				data, err := json.Marshal(want)
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(data, &got); err != nil {
					t.Fatal(err)
				}
			case "gob":
				var buf bytes.Buffer
				if err := gob.NewEncoder(&buf).Encode(want); err != nil {
					t.Fatal(err)
				}
				if err := gob.NewDecoder(&buf).Decode(&got); err != nil {
					t.Fatal(err)
				}
			case "xml":
				data, err := xml.Marshal(want)
				if err != nil {
					t.Fatal(err)
				}
				if err := xml.Unmarshal(data, &got); err != nil {
					t.Fatal(err)
				}
			}
			if got != want {
				t.Fatal("lost active unit variant")
			}
		})
	}
}
