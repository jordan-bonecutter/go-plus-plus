// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slog_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

type unionChoice union {
	Number int
	Text   string
}

func TestUnionJSON(t *testing.T) {
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("union", "choice", unionChoice.Number(42))
	if !strings.Contains(buf.String(), `"choice":{"Number":42}`) {
		t.Fatalf("missing tagged union: %s", &buf)
	}
}
