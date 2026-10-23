// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcweb_test

import (
	"cmd/go/internal/vcweb"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	s, err := vcweb.NewServer(os.DevNull, t.TempDir(), log.Default())
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(s)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/help")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatal(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", body)
}

func TestOverview(t *testing.T) {
	s, err := vcweb.NewServer(os.DevNull, t.TempDir(), log.Default())
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(s)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatal(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", body)
}

func TestBackgroundFailureCached(t *testing.T) {
	scriptDir := t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	// An unknown test flag makes the background child fail on every platform.
	content := "exec '" + strings.ReplaceAll(exe, "'", "''") + "' -vcweb-invalid-flag &\nhandle dir\n"
	if err := os.WriteFile(filepath.Join(scriptDir, "fail.txt"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := vcweb.NewServer(scriptDir, t.TempDir(), log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for range 2 {
		called := false
		err := s.HandleScript("fail.txt", log.New(io.Discard, "", 0), func(http.Handler) { called = true })
		if err == nil || called {
			t.Fatalf("background failure: error = %v, handler called = %v", err, called)
		}
	}
}
