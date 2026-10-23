// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary

import (
	"bytes"
	"testing"
)

type unionChoice union {
	Number uint64
	Again  uint64
}

func TestUnionRejected(t *testing.T) {
	value := unionChoice.Again(42)
	var buf bytes.Buffer
	if Size(value) != -1 || Size([]unionChoice{value}) != -1 || Size([0]unionChoice{}) != -1 {
		t.Fatal("union has a fixed binary size")
	}
	if err := Write(&buf, LittleEndian, value); err == nil {
		t.Fatal("wrote private union storage")
	}
	if _, err := Encode(make([]byte, 100), LittleEndian, value); err == nil {
		t.Fatal("encoded private union storage")
	}
	if _, err := Append(nil, LittleEndian, value); err == nil {
		t.Fatal("appended private union storage")
	}
	if err := Read(bytes.NewReader(make([]byte, 100)), LittleEndian, &value); err == nil {
		t.Fatal("read private union storage")
	}
	if _, err := Decode(make([]byte, 100), LittleEndian, &value); err == nil {
		t.Fatal("decoded private union storage")
	}
	if value != unionChoice.Again(42) {
		t.Fatal("modified rejected union")
	}
}
