package status_test

import (
	"net/http"
	"net/http/status/v2"
	"testing"
)

func TestResponseCodes(t *testing.T) {
	for code := 100; code < 600; code++ {
		status := status.Parse(code)
		if status.Int() != code {
			t.Fatal(code, status)
		}
		if status.Text().UnwrapOr("") != http.StatusText(code) {
			t.Fatal(code)
		}
	}
	if status.Code.OK().String() != "200 OK" {
		t.Fatal("OK")
	}
	if status.Parse(599).String() != "599" {
		t.Fatal("extension")
	}
}
