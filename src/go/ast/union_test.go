// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ast

import (
	"bytes"
	"go/token"
	"strings"
	"testing"
)

type unionChoice union {
	Number int
	hidden string
}

func TestPrintUnion(t *testing.T) {
	for _, value := range []unionChoice{unionChoice.Number(42), unionChoice.hidden("private")} {
		var buf bytes.Buffer
		if err := Fprint(&buf, token.NewFileSet(), value, nil); err != nil {
			t.Fatal(err)
		}
		text := buf.String()
		if strings.Contains(text, "union·") || strings.Contains(text, "private") {
			t.Fatalf("printed private storage: %s", text)
		}
		if value == unionChoice.Number(42) && !strings.Contains(text, "Number(42)") {
			t.Fatalf("missing payload: %s", text)
		}
	}
}
