#!/usr/bin/env bash
# Check installed compiler, vet, and coverage tools against postfix ! source.
set -euo pipefail
bang_root=$(cd "$(dirname "$0")/.." && pwd)
export GOROOT="$bang_root" GOTOOLCHAIN=local GOWORK=off
bang_test_dir=$(mktemp -d "${TMPDIR:-/tmp}/bang-toolchain.XXXXXXXX")
trap 'rm -rf "$bang_test_dir"' EXIT
cat > "$bang_test_dir/go.mod" <<'EOF'
module bangtooling

go 1.26
EOF
cat > "$bang_test_dir/main.go" <<'EOF'
package bangtooling
import "errors"
var boom = errors.New("boom")
func value(fail bool) (int, error) {
 if fail { return 99, boom }
 return 7, nil
}
func Read(fail bool) (int, error) { return value(fail)!, nil }
func seq(y func(int) bool) { y(1) }
func Named() (n int, e error) {
 n = 42
 defer func() { n++ }()
 for range seq { value(true)! }
 return
}
EOF
cat > "$bang_test_dir/main_test.go" <<'EOF'
package bangtooling
import "testing"
func TestRead(t *testing.T) {
 n,e:=Read(false);if n!=7||e!=nil{t.Fatal(n,e)}
 n,e=Read(true);if n!=0||e!=boom{t.Fatal(n,e)}
 n,e=Named();if n!=43||e!=boom{t.Fatal(n,e)}
}
EOF
cd "$bang_test_dir"
"$bang_root/bin/go" vet ./...
"$bang_root/bin/go" test -coverprofile=coverage.out ./...
"$bang_root/bin/go" tool cover -func=coverage.out
"$bang_root/bin/go" test -race ./...
