// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template_test

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

type unionKey union {
	Z string
	A string
}
type unionDynamic union{ Any any }

func TestUnion(t *testing.T) {
	tmpl := template.Must(template.New("union").Parse(`{{range $k, $v := .}}{{$v}}{{end}}`))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[unionKey]string{unionKey.Z("a"): "1", unionKey.Z("b"): "2", unionKey.A("a"): "3"}); err != nil || buf.String() != "123" {
		t.Fatalf("union map order: %s %v", &buf, err)
	}
	tmpl = template.Must(template.New("eq").Parse(`{{eq . .}}`))
	buf.Reset()
	if err := tmpl.Execute(&buf, unionKey.A("x")); err != nil || buf.String() != "true" {
		t.Fatalf("union equality: %s %v", &buf, err)
	}
	err := tmpl.Execute(&buf, unionDynamic.Any([]int{1}))
	if err == nil || !strings.Contains(err.Error(), "non-comparable type") {
		t.Fatalf("dynamic comparability: %v", err)
	}
}

func TestUnionTruth(t *testing.T) {
	for _, value := range []any{unionKey{}, unionKey.A("x"), new(unionKey)} {
		tmpl := template.Must(template.New("truth").Parse(`{{if .}}true{{else}}false{{end}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, value); err != nil || buf.String() != "true" {
			t.Fatalf("union truth: %s %v", &buf, err)
		}
	}
}
