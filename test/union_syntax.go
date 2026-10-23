// errorcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type Option[T] union { // ERROR "missing constraint for type parameter T .use .T any.."
 None struct{}
 Some(T)
}
