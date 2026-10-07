# Release process

Releases are cut from a tag by `scripts/release.sh`, which does the local
half (test everything, build the binaries, write the checksums) and lets CI
do the public half (attach the artifacts and announce the version). The rule
that governs code also governs releases: **a tag names a version whose full
test suite is green.**

## Versioning

The project reads plain semantic versions. A tag is `v1.2.3`; the script and
the CI workflow agree on that shape and refuse anything else.

## Cutting a release

From a working tree that is clean and whose tests pass:

```sh
git tag v0.9.0
git push origin v0.9.0
```

That's the whole ceremony. The tag fires `.github/workflows/release.yml`,
which builds the binary for linux, darwin and windows on amd64 and arm64,
writes SHA-256 checksums, and publishes them on the GitHub release. The local
script exists to let you *verify* the same artifacts before you push the tag:

```sh
scripts/release.sh v0.9.0
```

It runs the full suite, builds every target into `dist/`, and prints the
checksums. Nothing is published until you push the tag.

## Why this shape

- One-sentence rule: green tests, then ship. Nothing in `scripts/` or the
  workflow skips `go test ./...`.
- Artifacts are reproducible: the workflow pins the Go version and builds
  from the tag, so the checksums a maintainer sees locally match what GitHub
  publishes.
- A failed step fails the run. There is no "best effort" release.

## What changes a release number

- **patch** — a parity fix or a newly wired function, nothing that changes
  behavior wire-format.
- **minor** — a new stdlib package, a new CLI flag, a new verification rule.
- **major** — a change that breaks compiled output or drops an option.

The compiled JavaScript is output, and output can change shape quietly; only
the *Go* API of the tool follows semver as strictly as this.