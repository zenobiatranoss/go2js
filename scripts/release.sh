#!/bin/sh
# the local half of a release: test everything, build the artifact matrix into
# dist/, and print the checksums. the public half is the tag firing CI, which
# publishes those same binaries. a tag looks like v1.2.3 and nothing else.
set -eu
cd "$(dirname "$0")/.."
version="${1:?usage: scripts/release.sh <v1.2.3>}"

case "$version" in
v[0-9]*.[0-9]*.[0-9]*) ;;
*) echo "tag must look like v1.2.3" >&2; exit 1 ;;
esac

go test ./...

root="dist/go2js-$version"
rm -rf "$root"
mkdir -p "$root"

for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do
	os="${target%/*}"
	arch="${target#*/}"
	ext=""
	[ "$os" = windows ] && ext=".exe"
	out="$root/go2js-$os-$arch$ext"
	GOOS="$os" GOARCH="$arch" go build -o "$out" ./cmd/go2js
done

if command -v sha256sum > /dev/null 2>&1; then
	( cd "$root" && sha256sum go2js-* > SHA256SUMS )
else
	( cd "$root" && shasum -a 256 go2js-* > SHA256SUMS )
fi

echo "artifacts in $root:"
ls -1 "$root" | sed 's/^/  /'