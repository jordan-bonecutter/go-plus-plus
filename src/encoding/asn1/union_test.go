// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asn1

import "testing"

type unionChoice union {
	Number int
	Again  int
}

func TestUnionRejected(t *testing.T) {
	value := unionChoice.Again(42)
	if _, err := Marshal(value); err == nil {
		t.Fatal("marshaled private union storage")
	}
	if _, err := Marshal(struct{ Value unionChoice }{value}); err == nil {
		t.Fatal("marshaled nested union")
	}
	if _, err := Unmarshal([]byte{0x30, 0}, &value); err == nil {
		t.Fatal("unmarshaled private union storage")
	}
	if value != unionChoice.Again(42) {
		t.Fatal("modified rejected union")
	}
}
