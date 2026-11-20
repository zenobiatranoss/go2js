#!/usr/bin/env bash
# Rebuilds every example: runs the Go program, turns it into readable
# JavaScript, runs that under Node, and checks both print the same.
#
#   examples/build.sh              all examples
#   examples/build.sh 03-worker-pool
#
# Needs: go, node, bin/go2js and tools/jsmin/go2js-min (make -C tools/jsmin).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
export GO2JS_MIN="${GO2JS_MIN:-$root/tools/jsmin/go2js-min}"
go2js="${GO2JS:-$root/bin/go2js}"

if [ ! -x "$go2js" ]; then
	mkdir -p "$root/bin"
	(cd "$root" && go build -o bin/go2js ./cmd/go2js)
fi
if [ ! -x "$GO2JS_MIN" ]; then
	make -C "$root/tools/jsmin" >/dev/null
fi

cd "$root/examples"
failed=0
for dir in ${@:-[0-9]*/}; do
	dir="${dir%/}"
	(cd "$dir" && go run main.go > output.txt 2>&1)
	mkdir -p "$dir/js"
	"$go2js" -readable -module commonjs -o "$dir/js/main.js" "$dir/main.go"
	if node "$dir/js/main.js" 2>&1 | diff -q - "$dir/output.txt" >/dev/null; then
		printf '%-18s same output, %5d lines of JavaScript\n' "$dir" "$(wc -l < "$dir/js/main.js")"
	else
		printf '%-18s DIFFERENT\n' "$dir"
		failed=1
	fi
done
exit $failed
