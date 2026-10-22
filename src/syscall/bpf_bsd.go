// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

// Berkeley packet filter for BSD variants

package syscall

import (
	"unsafe"
)

// Deprecated: Use golang.org/x/net/bpf instead.
func BpfStmt(code, k int) *BpfInsn {
	return &BpfInsn{Code: uint16(code), K: uint32(k)}
}

// Deprecated: Use golang.org/x/net/bpf instead.
func BpfJump(code, k, jt, jf int) *BpfInsn {
	return &BpfInsn{Code: uint16(code), Jt: uint8(jt), Jf: uint8(jf), K: uint32(k)}
}

// Deprecated: Use golang.org/x/net/bpf instead.
func BpfBuflen(fd int) (int, error) {
	var l int
	ioctlPtr(fd, BIOCGBLEN, unsafe.Pointer(&l))!
	return l, nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func SetBpfBuflen(fd, l int) (int, error) {
	ioctlPtr(fd, BIOCSBLEN, unsafe.Pointer(&l))!
	return l, nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func BpfDatalink(fd int) (int, error) {
	var t int
	ioctlPtr(fd, BIOCGDLT, unsafe.Pointer(&t))!
	return t, nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func SetBpfDatalink(fd, t int) (int, error) {
	ioctlPtr(fd, BIOCSDLT, unsafe.Pointer(&t))!
	return t, nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func SetBpfPromisc(fd, m int) error {
	ioctlPtr(fd, BIOCPROMISC, unsafe.Pointer(&m))!
	return nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func FlushBpf(fd int) error {
	ioctlPtr(fd, BIOCFLUSH, nil)!
	return nil
}

type ivalue struct {
	name  [IFNAMSIZ]byte
	value int16
}

// Deprecated: Use golang.org/x/net/bpf instead.
func BpfInterface(fd int, name string) (string, error) {
	var iv ivalue
	ioctlPtr(fd, BIOCGETIF, unsafe.Pointer(&iv))!
	return name, nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func SetBpfInterface(fd int, name string) error {
	var iv ivalue
	copy(iv.name[:], []byte(name))
	ioctlPtr(fd, BIOCSETIF, unsafe.Pointer(&iv))!
	return nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func BpfTimeout(fd int) (*Timeval, error) {
	var tv Timeval
	ioctlPtr(fd, BIOCGRTIMEOUT, unsafe.Pointer(&tv))!
	return &tv, nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func SetBpfTimeout(fd int, tv *Timeval) error {
	ioctlPtr(fd, BIOCSRTIMEOUT, unsafe.Pointer(tv))!
	return nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func BpfStats(fd int) (*BpfStat, error) {
	var s BpfStat
	ioctlPtr(fd, BIOCGSTATS, unsafe.Pointer(&s))!
	return &s, nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func SetBpfImmediate(fd, m int) error {
	ioctlPtr(fd, BIOCIMMEDIATE, unsafe.Pointer(&m))!
	return nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func SetBpf(fd int, i []BpfInsn) error {
	var p BpfProgram
	p.Len = uint32(len(i))
	p.Insns = (*BpfInsn)(unsafe.Pointer(&i[0]))
	ioctlPtr(fd, BIOCSETF, unsafe.Pointer(&p))!
	return nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func CheckBpfVersion(fd int) error {
	var v BpfVersion
	ioctlPtr(fd, BIOCVERSION, unsafe.Pointer(&v))!
	if v.Major != BPF_MAJOR_VERSION || v.Minor != BPF_MINOR_VERSION {
		return EINVAL
	}
	return nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func BpfHeadercmpl(fd int) (int, error) {
	var f int
	ioctlPtr(fd, BIOCGHDRCMPLT, unsafe.Pointer(&f))!
	return f, nil
}

// Deprecated: Use golang.org/x/net/bpf instead.
func SetBpfHeadercmpl(fd, f int) error {
	ioctlPtr(fd, BIOCSHDRCMPLT, unsafe.Pointer(&f))!
	return nil
}
