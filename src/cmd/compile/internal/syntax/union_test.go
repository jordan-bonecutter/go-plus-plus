// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import "testing"

func TestUnionTagWidths(t *testing.T) {
	for _, tt := range []struct {
		count int
		name  string
	}{
		{1, "union·uint8"}, {256, "union·uint8"}, {257, "union·uint16"},
		{65536, "union·uint16"}, {65537, "union·uint32"},
	} {
		st := new(StructType)
		st.FieldList = append(st.FieldList, unionField(Pos{}, UnionTagField, nil))
		for i := 0; i < tt.count; i++ {
			st.FieldList = append(st.FieldList, unionField(Pos{}, UnionFieldPrefix+"V", NewName(Pos{}, "int")))
		}
		d := &TypeDecl{Type: st}
		lowerUnion(d)
		if got := st.FieldList[2].Type.(*Name).Value; got != tt.name {
			t.Errorf("%d variants: tag %s, want %s", tt.count, got, tt.name)
		}
	}
}
