// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"net"
	"net/rpc"
	"testing"
)

type UnionArgument union {
	Number int
	Text   string
}
type UnionService struct{}

func (*UnionService) Echo(arg UnionArgument, reply *UnionArgument) error { *reply = arg; return nil }

func TestUnionRPC(t *testing.T) {
	server := rpc.NewServer()
	if err := server.Register(new(UnionService)); err != nil {
		t.Fatal(err)
	}
	peer, conn := net.Pipe()
	defer peer.Close()
	go server.ServeConn(peer)
	client := rpc.NewClient(conn)
	defer client.Close()
	for _, want := range []UnionArgument{UnionArgument.Number(0), UnionArgument.Text("hello")} {
		var got UnionArgument
		if err := client.Call("UnionService.Echo", want, &got); err != nil || got != want {
			t.Fatalf("union RPC: %v %v", got, err)
		}
	}
}
