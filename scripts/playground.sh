#!/bin/sh
# builds the playground: the compiler as WebAssembly, beside the page that runs
# it. Serve the playground directory with any static server to use it.
set -eu
cd "$(dirname "$0")/.."

goroot="$(go env GOROOT)"

GOOS=js GOARCH=wasm go build -o playground/playground.wasm ./cmd/playground
cp "$goroot/lib/wasm/wasm_exec.js" playground/wasm_exec.js
cp go2jslog.png playground/go2jslog.png

# the shrinker of the page is go2js-min built as WebAssembly with the wasi-sdk,
# which WASI_SDK_PATH names; without it the page runs, with the shrinking off
if [ -n "${WASI_SDK_PATH:-}" ]; then
	"$WASI_SDK_PATH/bin/clang++" --sysroot="$WASI_SDK_PATH/share/wasi-sysroot" -O2 -std=c++17 \
		-fno-exceptions -mexec-model=reactor -Wl,--strip-all -o playground/go2js-min.wasm \
		tools/jsmin/lx.cpp tools/jsmin/em.cpp tools/jsmin/rn.cpp tools/jsmin/fm.cpp tools/jsmin/pg.cpp tools/jsmin/wa.cpp
	echo "playground/go2js-min.wasm"
fi

echo "playground/playground.wasm"
