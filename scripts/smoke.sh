#!/bin/sh
# compiles and runs each named example twice — once with go run, once through
# go2js and node — and refuses to pass on any difference. the examples are
# named by their directory under examples/; with no names, every example runs.
set -eu
cd "$(dirname "$0")/.."
"$(dirname "$0")/build.sh"

if [ "$#" -eq 0 ]; then
	set -- hello_world async_channels http_shapes realworld
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

for ex in "$@"; do
	dir="examples/$ex"
	bin/go2js -o "$tmp/out.js" "$dir"

	go run "./$dir" > "$tmp/go.out"
	node "$tmp/out.js" > "$tmp/js.out"

	if ! diff -q "$tmp/go.out" "$tmp/js.out" > /dev/null; then
		echo "smoke: $ex differs between go run and node:" >&2
		diff "$tmp/go.out" "$tmp/js.out" >&2 || true
		exit 1
	fi

	echo "smoke ok: $ex"
done