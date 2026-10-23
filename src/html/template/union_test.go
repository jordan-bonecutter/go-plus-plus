// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template_test

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

type unionChoice union {
	Number int
	Text   string
}

func TestUnionJavaScript(t *testing.T) {
	tmpl := template.Must(template.New("union").Parse(`<script>const choice = {{.}};</script>`))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, unionChoice.Number(42)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `{"Number":42}`) {
		t.Fatalf("missing tagged union: %s", &buf)
	}
}
