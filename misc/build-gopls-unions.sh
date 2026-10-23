#!/usr/bin/env bash
# Build the union-aware editor server with this checkout's parser and checker.
set -euo pipefail

union_root=$(cd "$(dirname "$0")/.." && pwd)
union_go="$union_root/bin/go"
union_version=v0.23.0
export GOTOOLCHAIN=local
export GOWORK=off

"$union_go" mod download "golang.org/x/tools/gopls@$union_version"
union_cache=$("$union_go" env GOMODCACHE)
union_build=$(mktemp -d "${TMPDIR:-/tmp}/gopls-unions.XXXXXXXX")
trap 'rm -rf "$union_build"' EXIT
cp -R "$union_cache/golang.org/x/tools/gopls@$union_version/." "$union_build/"
chmod -R u+w "$union_build"
patch -s -d "$union_build" -p1 < "$union_root/misc/gopls-unions.patch"
cd "$union_build"
"$union_go" build -buildvcs=false -mod=readonly -o "$union_root/bin/gopls" .
printf 'Built %s/bin/gopls\n' "$union_root"
