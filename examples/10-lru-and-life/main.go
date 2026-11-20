// LRU cache with container/list, Conway's game of life and a seeded maze via a custom RNG.
package main

import (
	"container/list"
	"fmt"
	"strings"
)

type LRU struct {
	cap          int
	order        *list.List
	items        map[string]*list.Element
	hits, misses int
}

type kv struct {
	key string
	val int
}

func NewLRU(n int) *LRU {
	return &LRU{cap: n, order: list.New(), items: map[string]*list.Element{}}
}

func (c *LRU) Get(k string) (int, bool) {
	if e, ok := c.items[k]; ok {
		c.order.MoveToFront(e)
		c.hits++
		return e.Value.(*kv).val, true
	}
	c.misses++
	return 0, false
}

func (c *LRU) Put(k string, v int) (evicted string) {
	if e, ok := c.items[k]; ok {
		e.Value.(*kv).val = v
		c.order.MoveToFront(e)
		return ""
	}
	c.items[k] = c.order.PushFront(&kv{k, v})
	if c.order.Len() > c.cap {
		last := c.order.Back()
		c.order.Remove(last)
		delete(c.items, last.Value.(*kv).key)
		return last.Value.(*kv).key
	}
	return ""
}

func (c *LRU) Keys() []string {
	var out []string
	for e := c.order.Front(); e != nil; e = e.Next() {
		out = append(out, e.Value.(*kv).key)
	}
	return out
}

type Grid [][]bool

func ParseGrid(rows ...string) Grid {
	g := make(Grid, len(rows))
	for y, r := range rows {
		g[y] = make([]bool, len(r))
		for x, c := range r {
			g[y][x] = c == '#'
		}
	}
	return g
}

func (g Grid) Neighbours(x, y int) int {
	n := 0
	h, w := len(g), len(g[0])
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if (dx != 0 || dy != 0) && g[(y+dy+h)%h][(x+dx+w)%w] {
				n++
			}
		}
	}
	return n
}

func (g Grid) Step() Grid {
	next := make(Grid, len(g))
	for y := range g {
		next[y] = make([]bool, len(g[y]))
		for x := range g[y] {
			n := g.Neighbours(x, y)
			next[y][x] = n == 3 || (g[y][x] && n == 2)
		}
	}
	return next
}

func (g Grid) String() string {
	var sb strings.Builder
	for _, row := range g {
		for _, c := range row {
			if c {
				sb.WriteByte('#')
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

type RNG struct{ state uint32 }

func (r *RNG) Next() uint32 {
	r.state = r.state*1664525 + 1013904223
	return r.state >> 8
}
func (r *RNG) Intn(n int) int { return int(r.Next() % uint32(n)) }

func maze(w, h int, rng *RNG) []string {
	grid := make([][]byte, h*2+1)
	for y := range grid {
		grid[y] = []byte(strings.Repeat("#", w*2+1))
	}
	seen := make([][]bool, h)
	for i := range seen {
		seen[i] = make([]bool, w)
	}
	type cell struct{ x, y int }
	stack := []cell{{0, 0}}
	seen[0][0] = true
	grid[1][1] = ' '
	for len(stack) > 0 {
		c := stack[len(stack)-1]
		var opts []cell
		for _, d := range []cell{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := c.x+d.x, c.y+d.y
			if nx >= 0 && ny >= 0 && nx < w && ny < h && !seen[ny][nx] {
				opts = append(opts, cell{nx, ny})
			}
		}
		if len(opts) == 0 {
			stack = stack[:len(stack)-1]
			continue
		}
		n := opts[rng.Intn(len(opts))]
		seen[n.y][n.x] = true
		grid[n.y*2+1][n.x*2+1] = ' '
		grid[c.y+n.y+1][c.x+n.x+1] = ' '
		stack = append(stack, n)
	}
	out := make([]string, len(grid))
	for i, row := range grid {
		out[i] = string(row)
	}
	return out
}

func main() {
	c := NewLRU(3)
	for _, k := range []string{"a", "b", "c", "a", "d", "b", "e", "a"} {
		if _, ok := c.Get(k); !ok {
			if ev := c.Put(k, len(k)); ev != "" {
				fmt.Printf("put %s evicts %s\n", k, ev)
			}
		}
	}
	fmt.Println(c.Keys(), "hits", c.hits, "misses", c.misses)

	g := ParseGrid(
		"..........",
		"...#......",
		"....#.....",
		"..###.....",
		"..........",
		"..........",
	)
	for gen := 0; gen < 3; gen++ {
		fmt.Printf("generation %d\n%s", gen, g)
		g = g.Step()
		g = g.Step()
	}

	rng := &RNG{42}
	for _, line := range maze(10, 5, rng) {
		fmt.Println(line)
	}
	fmt.Println(rng.Next(), rng.Intn(100))
}
