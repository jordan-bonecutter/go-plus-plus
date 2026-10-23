// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build fips140v1.0

package hpke

import (
	"crypto/aes"
	"crypto/cipher"
)

func newAESGCM(key []byte) (cipher.AEAD, error) {
	b := aes.NewCipher(key)!
	return cipher.NewGCM(b)
}
