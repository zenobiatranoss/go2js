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
| uuid, axios, zod, Howler, p5, Chart.js | declarations only |
| semver | not read: re-exports through `import x = require` |
| lodash | not read: `/// <reference path>` |

Known gaps: other mapped types (`Pick`, `keyof`-based) become `any`; `Partial`, `Readonly`, `Required` and `Omit` are read as the type itself; a `Promise` as a
parameter, property or callback result is not supported; pointers to bound structs
are not supported (value structs only).

Platforms: the tests run on Linux here. macOS and Windows run them in CI
(`.github/workflows/portable.yml`); they were not run on those systems in this
session. The headless Chromium check is manual.
