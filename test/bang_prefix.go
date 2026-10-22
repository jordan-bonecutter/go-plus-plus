// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Preserve the newlines after prefix ! to exercise semicolon insertion.

func negated() bool {
	return !
		false
}
func commented() bool {
	return ! /*
	 comment */ false
}
func nested() bool {
	return ! !
		true
}
func only() error {	return nil }
func postfix() error {
	only() /* comment */ !
	return nil
}
func main() {
	if !negated() || !commented() || !nested() || postfix() != nil {
		panic("negation/propagation semicolons")
	}
}
