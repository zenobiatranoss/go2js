#!/bin/bash
# Compile every benchmark in this folder with go2js and run it under node,
# saving a cpu profile for each under profiles/, then run the same program
# with go as a baseline. The profiles open in the flame chart of Chrome.
set -eu
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"
GO2JS="${GO2JS:-$ROOT/bin/go2js}"
OUT="$HERE/out"
PROF="$HERE/profiles"
mkdir -p "$OUT" "$PROF"

for dir in "$HERE"/*/; do
	name=$(basename "$dir")
	if [ ! -f "$dir/main.go" ]; then
		continue
	fi

	echo "== $name =="
	echo "-- go baseline --"
	(cd "$dir" && go run .)
	echo "-- go2js --"
	"$GO2JS" -o "$OUT/$name.js" "$dir"
	node --cpu-prof --cpu-prof-dir="$PROF" --cpu-prof-name="$name.cpuprofile" "$OUT/$name.js"
	echo "profile: $PROF/$name.cpuprofile"
done