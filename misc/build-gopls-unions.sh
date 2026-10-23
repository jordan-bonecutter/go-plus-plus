#!/usr/bin/env bash
# Build the union- and bang-aware editor server with this checkout's toolchain.
set -euo pipefail

union_root=$(cd "$(dirname "$0")/.." && pwd)
union_go="$union_root/bin/go"
union_version=v0.23.0
union_tools_version=v0.47.1-0.20260707181000-a299dadba899
union_staticcheck_version=v0.8.0-rc.1
export GOROOT="$union_root"
export GOTOOLCHAIN=local
export GOWORK=off

"$union_go" mod download "golang.org/x/tools/gopls@$union_version" \
  "golang.org/x/tools@$union_tools_version" "honnef.co/go/tools@$union_staticcheck_version"
union_cache=$("$union_go" env GOMODCACHE)
union_work=$(mktemp -d "${TMPDIR:-/tmp}/gopls-unions.XXXXXXXX")
trap 'rm -rf "$union_work"' EXIT
union_build="$union_work/gopls"
union_tools="$union_work/tools"
union_staticcheck="$union_work/staticcheck"
mkdir "$union_build" "$union_tools" "$union_staticcheck"
cp -R "$union_cache/golang.org/x/tools/gopls@$union_version/." "$union_build/"
cp -R "$union_cache/golang.org/x/tools@$union_tools_version/." "$union_tools/"
cp -R "$union_cache/honnef.co/go/tools@$union_staticcheck_version/." "$union_staticcheck/"
chmod -R u+w "$union_work"
patch -s -d "$union_build" -p1 < "$union_root/misc/gopls-unions.patch"
patch -s -d "$union_build" -p1 < "$union_root/misc/gopls-union-navigation.patch"
patch -s -d "$union_build" -p1 < "$union_root/misc/gopls-bang-refactor.patch"
patch -s -d "$union_tools" -p1 < "$union_root/misc/gopls-bang-tools.patch"
patch -s -d "$union_staticcheck" -p1 < "$union_root/misc/gopls-bang-staticcheck.patch"
cd "$union_build"
"$union_go" mod edit "-replace=golang.org/x/tools=$union_tools" "-replace=honnef.co/go/tools=$union_staticcheck"
"$union_go" run -mod=readonly "$union_root/misc/test-gopls-bang-analysis.go"
"$union_go" build -buildvcs=false -mod=readonly -o "$union_root/bin/gopls" .
printf 'Built %s/bin/gopls\n' "$union_root"
