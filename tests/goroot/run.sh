#!/bin/bash
# The experiment of compiling real GOROOT packages: each probe under probes/
# imports packages from the Go installation itself, and this script asks both
# Go and go2js to build and run it, then reports which of the two answered
# differently. A probe that go2js can't compile yet is the honest status of
# that package's port, not a silent hole.
set -eu
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/../.." && pwd)"
GO2JS="${GO2JS:-$ROOT/bin/go2js}"
tmp_go="$HERE/.build/go.out"
tmp_js="$HERE/.build/js.out"
tmp_out="$HERE/.build/out.js"
mkdir -p "$HERE/.build"

pass=0
fail=0
for dir in "$HERE"/probes/*/; do
	name=$(basename "$dir")
	go_code=0
	js_code=0
	(cd "$dir" && go run .) > "$tmp_go" 2>&1 || go_code=$?

	if ! "$GO2JS" -o "$tmp_out" "$dir" > /dev/null 2>&1; then
		echo "COMPILE-FAIL $name (go2js can't build it yet)"
		fail=$((fail + 1))
		continue
	fi
	node "$tmp_out" > "$tmp_js" 2>&1 || js_code=$?

	if [ "$go_code" -ne "$js_code" ] || ! diff -q "$tmp_go" "$tmp_js" > /dev/null 2>&1; then
		echo "DIFF $name"
		diff "$tmp_go" "$tmp_js" | head -16 || true
		fail=$((fail + 1))
		continue
	fi
	echo "OK   $name"
	pass=$((pass + 1))
done
rm -rf "$HERE/.build"
echo "$pass pass, $fail fail"
[ "$fail" -eq 0 ]