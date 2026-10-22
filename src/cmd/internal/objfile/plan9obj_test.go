// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objfile

import (
	"bytes"
	"debug/plan9obj"
	"encoding/binary"
	"testing"
)

func TestPlan9TableBounds(t *testing.T) {
	const textStart = 0x1020
	for _, tt := range []struct {
		name        string
		start, end  uint32
		missingText bool
		wantErr     bool
	}{
		{"valid", textStart, textStart + 4, false, false},
		{"missing_text", textStart, textStart + 4, true, true},
		{"before_text", textStart - 1, textStart + 4, false, true},
		{"reversed", textStart + 3, textStart + 2, false, true},
		{"past_text", textStart, textStart + 5, false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var syms []byte
			for _, s := range []struct {
				name  string
				value uint32
			}{{"start", tt.start}, {"end", tt.end}} {
				syms = binary.BigEndian.AppendUint32(syms, s.value)
				syms = append(syms, 'T')
				syms = append(syms, s.name...)
				syms = append(syms, 0)
			}
			var image []byte
			for _, word := range []uint32{plan9obj.Magic386, 4, 0, 0, uint32(len(syms)), 0, 0, 0} {
				image = binary.BigEndian.AppendUint32(image, word)
			}
			image = append(append(image, 1, 2, 3, 4), syms...)
			f, err := plan9obj.NewFile(bytes.NewReader(image))
			if err != nil {
				t.Fatal(err)
			}
			if tt.missingText {
				f.Sections = f.Sections[1:]
			}
			data, err := loadPlan9Table(f, "start", "end")
			if (err != nil) != tt.wantErr {
				t.Fatalf("loadPlan9Table = %v, %v; want error %v", data, err, tt.wantErr)
			}
			if !tt.wantErr && !bytes.Equal(data, []byte{1, 2, 3, 4}) {
				t.Fatalf("table = %v", data)
			}
		})
	}
}
