#!/usr/bin/env bash
# Build this checkout and install go++, go++pls, and go++fmt launchers.
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
config_home=${GOPP_CONFIG_HOME:-${XDG_CONFIG_HOME:-$HOME/.config}/go++}
cache_home=${GOPP_CACHE_HOME:-${XDG_CACHE_HOME:-$HOME/.cache}/go++}
install_bin=${GOPP_INSTALL_BIN:-$HOME/.local/bin}

# Bootstrap with a stock Go installation, never with this checkout's compiler.
if [[ -z ${GOROOT_BOOTSTRAP:-} ]]; then
  if [[ -x /usr/lib/go/bin/go ]]; then
    export GOROOT_BOOTSTRAP=/usr/lib/go
  else
    GOROOT_BOOTSTRAP=$(GOTOOLCHAIN=local go env GOROOT)
    export GOROOT_BOOTSTRAP
  fi
fi
if [[ $(cd "$GOROOT_BOOTSTRAP" && pwd -P) == "$root" ]]; then
  echo 'Set GOROOT_BOOTSTRAP to a stock Go installation.' >&2
  exit 1
fi

mkdir -p "$cache_home/build" "$cache_home/tmp"
export GOCACHE="$cache_home/build" TMPDIR="$cache_home/tmp"
export GOMODCACHE="$cache_home/mod" GOTOOLCHAIN=local
export GOFLAGS=-buildvcs=false GIT_OPTIONAL_LOCKS=0
(
  cd "$root/src"
  bash make.bash
)
bash "$root/misc/build-gopls-unions.sh"

# Install only after both builds succeed. Launchers follow this checkout.
mkdir -p "$config_home" "$install_bin"
{
  printf '# Shared environment for the Go++ command-line tools.\n'
  cat <<'ENV'
export GOPP_CACHE_HOME="${GOPP_CACHE_HOME:-${XDG_CACHE_HOME:-$HOME/.cache}/go++}"
export GOPP_CONFIG_HOME="${GOPP_CONFIG_HOME:-${XDG_CONFIG_HOME:-$HOME/.config}/go++}"
export GOPP_DATA_HOME="${GOPP_DATA_HOME:-${XDG_DATA_HOME:-$HOME/.local/share}/go++}"
export XDG_CACHE_HOME="$GOPP_CACHE_HOME"
export XDG_CONFIG_HOME="$GOPP_CONFIG_HOME"
export XDG_DATA_HOME="$GOPP_DATA_HOME"
ENV
  printf 'export GOROOT=%q\n' "$root"
  cat <<'ENV'
export GOTOOLCHAIN=local
export GOCACHE="$GOPP_CACHE_HOME/build"
export GOMODCACHE="$GOPP_CACHE_HOME/mod"
export GOPLSCACHE="$GOPP_CACHE_HOME"
export GOPATH="$GOPP_DATA_HOME"
export GOBIN="$GOPP_DATA_HOME/bin"
export GOENV="$GOPP_CONFIG_HOME/env"
export PATH="$GOROOT/bin:$GOBIN:$PATH"
mkdir -p "$GOPP_CACHE_HOME" "$GOPP_CONFIG_HOME" "$GOBIN"
ENV
} > "$config_home/toolchain.sh"

for pair in 'go++:go' 'go++pls:gopls' 'go++fmt:gofmt'; do
  name=${pair%%:*}
  executable=${pair#*:}
  {
    printf '#!/usr/bin/env bash\n'
    printf 'source %q\n' "$config_home/toolchain.sh"
    printf 'exec "$GOROOT/bin/%s" "$@"\n' "$executable"
  } > "$install_bin/$name"
  chmod +x "$install_bin/$name"
done
printf 'Installed go++, go++pls, and go++fmt in %s\n' "$install_bin"
