// errorcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func value() (int, error)     { return 0, nil }
func only() error             { return nil }
func nothing()                {}
func number() int             { return 0 }
func noError()                { only()! }                // ERROR "postfix ! requires enclosing function to return a final error"
func wrongLast() (error, int) { only()!; return nil, 0 } // ERROR "postfix ! requires enclosing function to return a final error"
var outside = value()!        // ERROR "postfix ! requires enclosing function to return a final error"
func invalid() error {
	nothing()!    // ERROR "postfix ! requires call to return a final error"
	number()!     // ERROR "postfix ! requires call to return a final error"
	len("")!      // ERROR "postfix ! requires an error-returning function call"
	int(0)!       // ERROR "postfix ! requires an error-returning function call"
	go only()!    // ERROR "postfix ! is not allowed in go call"
	defer only()! // ERROR "postfix ! is not allowed in defer call"
	return nil
}

type custom interface{ Error() string }

func customResult() custom   { return nil }
func wrongCall() error       { customResult()!; return nil } // ERROR "postfix ! requires call to return a final error"
func wrongEnclosing() custom { only()!; return nil }         // ERROR "postfix ! requires enclosing function to return a final error"
func wrongClosure() error {
	_ = func() { only()! } // ERROR "postfix ! requires enclosing function to return a final error"
	return nil
}
