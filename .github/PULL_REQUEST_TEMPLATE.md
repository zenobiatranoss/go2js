## What changed

A one-sentence behavior description, the way commit messages are written here.

## How it's proven

The parity rule is the law: a Go program must print the same thing when run
with `go run` and when compiled by go2js and run with node.

- [ ] a parity test accompanies the change (`tests/integration/`)
- [ ] `./scripts/check.sh` is green (gofmt, vet, build, full test suite)
- [ ] the examples still pass `./scripts/smoke.sh`, if the stdlib moved

## If the change is stdlib

- the function is registered in the right map in
  `backend/javascript/stdlib.go` (or lives under `packageTypes`/constants),
- its JS helper is in `backend/javascript/*.go` with a comment in the same
  voice, and
- `docs/stdlib-support.md` was regenerated:
  `go run ./tools/stdlib-gaps -write docs/stdlib-support.md`.