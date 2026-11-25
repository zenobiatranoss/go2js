# What `go2js bind` has been tried on

Only what was run is marked as working. "Declarations only" means the binding is
generated and compiles, but no program was run against the library.

| Library | Result |
|---|---|
| Browser (`-with dom`): canvas 2D, keyboard events, `requestAnimationFrame` | works, run in headless Chromium ([examples/browser-game](../examples/browser-game)) |
| Browser: WebGL (triangle drawn, buffer overloads) | works in headless Chromium (software GL) |
| three.js | works through an import map |
| Matter.js | works |
| dayjs, nanoid | work |
| uuid (v4, v5, validate), zod (string, number, parse), axios (typed response, failed request, against a local server) | work, run in Node |
| Howler, p5, Chart.js | declarations only |
| semver (41 functions, `SemVer` class) | works, run in Node: valid, gt, satisfies, parse, coerce, diff, clean |
| lodash (295 functions) | works, run in Node: chunk, uniq, sum, map, camelCase; its types are mostly `any` (see below) |

Known gaps: mapped types written with `keyof` or `in` become `any`; `Partial`, `Readonly`, `Required`, `Omit` and `Pick` are read as the type itself; generic collections (lodash) are `[]any`, and a whole number that JavaScript hands back inside an `any` is an `int` in Go (`v.(int)`, not `v.(float64)`); a `Promise` held in a
property is left out (a parameter or a callback result is its value); pointers to bound structs
are not supported (value structs only).

Platforms: the tests run on Linux here. macOS and Windows run them in CI
(`.github/workflows/portable.yml`); they were not run on those systems in this
session. The headless Chromium check is manual.
