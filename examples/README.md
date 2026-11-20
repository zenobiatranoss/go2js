# Examples

Ten complete Go programs, each next to the JavaScript go2js makes from it, so you can read
both. Every program prints something deterministic; the output of the real Go program is
kept in `output.txt`, and the JavaScript prints exactly the same thing.

| Program | What it shows |
|---|---|
| [01-inventory](01-inventory) | structs, interfaces, embedding, type switches, `sort.Slice`, formatted reports |
| [02-bank](02-bank) | custom errors, `errors.Is` / `errors.As` / `errors.Join`, `%w`, panic and recover, named results |
| [03-worker-pool](03-worker-pool) | goroutines, channels, `WaitGroup`, `Mutex`, pipelines, `select`, tickers and timeouts |
| [04-word-stats](04-word-stats) | `bufio.Scanner`, `strings`, `unicode`, maps, sorting, `strings.Builder` |
| [05-generics](05-generics) | constraints, generic functions and containers, `cmp`, `slices` |
| [06-calculator](06-calculator) | a lexer, a recursive-descent parser and an AST evaluator behind an interface |
| [07-graphs](07-graphs) | BFS, Dijkstra with `container/heap`, topological sort, union-find |
| [08-numbers](08-numbers) | a prime sieve, `math/big`, matrices, integer overflow, `math`, `%` verbs |
| [09-json-library](09-json-library) | `encoding/json` with tags, `text/tabwriter`, `time` |
| [10-lru-and-life](10-lru-and-life) | an LRU cache on `container/list`, Conway's life, a seeded maze |

Each folder holds `main.go`, `output.txt` (what Go prints), and `js/main.js` with
`js/main.runtime.js` (the program and the runtime it needs, as readable JavaScript).
Run one with `node examples/03-worker-pool/js/main.js`.

To rebuild all of them and check Go and JavaScript still agree:

```
make -C tools/jsmin && go build -o bin/go2js ./cmd/go2js && examples/build.sh
```

Writing these found real bugs in go2js (a `return` that lost a shadowed named result,
a struct with a method called `next` run as if it were an iterator, `%v` of a Stringer in
`Errorf`, `%f` of a whole float written to a `Fprintf` writer, `big.Int.MulRange`); they are
fixed and covered by the parity tests. One thing the examples keep clear of on purpose:
`encoding/json` does not yet read or write a `time.Time` field, so example 09 keeps its
dates as strings and parses them with `time.Parse`.
