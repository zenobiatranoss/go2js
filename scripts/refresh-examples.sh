#!/bin/sh
# regenerates the compiled main.js next to every example's source, so the
# committed output keeps matching the committed sources. the rule says no
# generated files by hand; this script is the hand.
set -eu
cd "$(dirname "$0")/.."
"$(dirname "$0")/build.sh"

for main in examples/*/main.go; do
	dir="${main%/main.go}"
	if bin/go2js -o "$dir/main.js" "$dir" 2> "${dir%/}/.refresherr"; then
		echo "refreshed $dir/main.js"
	else
		echo "failed $dir/main.go:" >&2
		cat "${dir%/}/.refresherr" >&2
		rm -f "${dir%/}/.refresherr"
		exit 1
	fi
	rm -f "${dir%/}/.refresherr"
done