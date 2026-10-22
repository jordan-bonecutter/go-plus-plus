// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"context"
	"io"
	"os"
)

func (c *TCPConn) readFrom(r io.Reader) (int64, error) {
	return genericReadFrom(c, r)
}

func (c *TCPConn) writeTo(w io.Writer) (int64, error) {
	return genericWriteTo(c, w)
}

func (sd *sysDialer) dialTCP(ctx context.Context, laddr, raddr *TCPAddr) (*TCPConn, error) {
	if h := sd.testHookDialTCP; h != nil {
		return h(ctx, sd.network, laddr, raddr)
	}
	if h := testHookDialTCP; h != nil {
		return h(ctx, sd.network, laddr, raddr)
	}
	return sd.doDialTCP(ctx, laddr, raddr)
}

func (sd *sysDialer) doDialTCP(ctx context.Context, laddr, raddr *TCPAddr) (*TCPConn, error) {
	switch sd.network {
	case "tcp4":
		// Plan 9 doesn't complain about [::]:0->127.0.0.1, so it's up to us.
		if laddr != nil && len(laddr.IP) != 0 && laddr.IP.To4() == nil {
			return nil, &AddrError{Err: "non-IPv4 local address", Addr: laddr.String()}
		}
	case "tcp", "tcp6":
	default:
		return nil, UnknownNetworkError(sd.network)
	}
	if raddr == nil {
		return nil, errMissingAddress
	}
	fd := dialPlan9(ctx, sd.network, laddr, raddr)!
	return newTCPConn(fd, sd.Dialer.KeepAlive, sd.Dialer.KeepAliveConfig, testPreHookSetKeepAlive, testHookSetKeepAlive), nil
}

func (ln *TCPListener) ok() bool { return ln != nil && ln.fd != nil && ln.fd.ctl != nil }

func (ln *TCPListener) accept() (*TCPConn, error) {
	fd := ln.fd.acceptPlan9()!
	return newTCPConn(fd, ln.lc.KeepAlive, ln.lc.KeepAliveConfig, testPreHookSetKeepAlive, testHookSetKeepAlive), nil
}

func (ln *TCPListener) close() error {
	ln.fd.pfd.Close()!
	if _, err := ln.fd.ctl.WriteString("hangup"); err != nil {
		ln.fd.ctl.Close()
		return err
	}
	ln.fd.ctl.Close()!
	return nil
}

func (ln *TCPListener) file() (*os.File, error) {
	f := ln.dup()!
	return f, nil
}

func (sl *sysListener) listenTCP(ctx context.Context, laddr *TCPAddr) (*TCPListener, error) {
	fd := listenPlan9(ctx, sl.network, laddr)!
	return &TCPListener{fd: fd, lc: sl.ListenConfig}, nil
}
