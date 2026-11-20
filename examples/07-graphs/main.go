// Graph algorithms: BFS, Dijkstra with container/heap, topological sort, union-find.
package main

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
	"strings"
)

type Edge struct {
	To     string
	Weight int
}

type Graph struct {
	adj map[string][]Edge
}

func NewGraph() *Graph { return &Graph{adj: map[string][]Edge{}} }

func (g *Graph) AddEdge(a, b string, w int, directed bool) {
	g.adj[a] = append(g.adj[a], Edge{b, w})
	if _, ok := g.adj[b]; !ok {
		g.adj[b] = nil
	}
	if !directed {
		g.adj[b] = append(g.adj[b], Edge{a, w})
	}
}

func (g *Graph) Nodes() []string {
	out := make([]string, 0, len(g.adj))
	for n := range g.adj {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func (g *Graph) BFS(start string) []string {
	seen := map[string]bool{start: true}
	queue := []string{start}
	var order []string
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		order = append(order, n)
		next := append([]Edge(nil), g.adj[n]...)
		sort.Slice(next, func(i, j int) bool { return next[i].To < next[j].To })
		for _, e := range next {
			if !seen[e.To] {
				seen[e.To] = true
				queue = append(queue, e.To)
			}
		}
	}
	return order
}

type entry struct {
	node string
	dist int
}
type pq []entry

func (p pq) Len() int { return len(p) }
func (p pq) Less(i, j int) bool {
	return p[i].dist < p[j].dist || (p[i].dist == p[j].dist && p[i].node < p[j].node)
}
func (p pq) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p *pq) Push(x any)   { *p = append(*p, x.(entry)) }
func (p *pq) Pop() any {
	old := *p
	n := len(old)
	x := old[n-1]
	*p = old[:n-1]
	return x
}

func (g *Graph) Dijkstra(src string) (map[string]int, map[string]string) {
	dist := map[string]int{}
	prev := map[string]string{}
	for n := range g.adj {
		dist[n] = math.MaxInt
	}
	dist[src] = 0
	h := &pq{{src, 0}}
	for h.Len() > 0 {
		cur := heap.Pop(h).(entry)
		if cur.dist > dist[cur.node] {
			continue
		}
		for _, e := range g.adj[cur.node] {
			if nd := cur.dist + e.Weight; nd < dist[e.To] {
				dist[e.To] = nd
				prev[e.To] = cur.node
				heap.Push(h, entry{e.To, nd})
			}
		}
	}
	return dist, prev
}

func path(prev map[string]string, to string) string {
	var p []string
	for n := to; n != ""; n = prev[n] {
		p = append([]string{n}, p...)
	}
	return strings.Join(p, " -> ")
}

func (g *Graph) TopoSort() ([]string, error) {
	indeg := map[string]int{}
	for n, es := range g.adj {
		if _, ok := indeg[n]; !ok {
			indeg[n] = 0
		}
		for _, e := range es {
			indeg[e.To]++
		}
	}
	var ready []string
	for n, d := range indeg {
		if d == 0 {
			ready = append(ready, n)
		}
	}
	sort.Strings(ready)
	var order []string
	for len(ready) > 0 {
		n := ready[0]
		ready = ready[1:]
		order = append(order, n)
		for _, e := range g.adj[n] {
			indeg[e.To]--
			if indeg[e.To] == 0 {
				ready = append(ready, e.To)
				sort.Strings(ready)
			}
		}
	}
	if len(order) != len(g.adj) {
		return nil, fmt.Errorf("cycle detected after %d of %d nodes", len(order), len(g.adj))
	}
	return order, nil
}

type UnionFind struct{ parent, size []int }

func NewUF(n int) *UnionFind {
	u := &UnionFind{make([]int, n), make([]int, n)}
	for i := range u.parent {
		u.parent[i], u.size[i] = i, 1
	}
	return u
}
func (u *UnionFind) Find(x int) int {
	for u.parent[x] != x {
		u.parent[x] = u.parent[u.parent[x]]
		x = u.parent[x]
	}
	return x
}
func (u *UnionFind) Union(a, b int) bool {
	a, b = u.Find(a), u.Find(b)
	if a == b {
		return false
	}
	if u.size[a] < u.size[b] {
		a, b = b, a
	}
	u.parent[b] = a
	u.size[a] += u.size[b]
	return true
}

func main() {
	g := NewGraph()
	for _, e := range []struct {
		a, b string
		w    int
	}{{"A", "B", 7}, {"A", "C", 9}, {"A", "F", 14}, {"B", "C", 10}, {"B", "D", 15}, {"C", "D", 11}, {"C", "F", 2}, {"D", "E", 6}, {"E", "F", 9}} {
		g.AddEdge(e.a, e.b, e.w, false)
	}
	fmt.Println("BFS from A:", g.BFS("A"))
	dist, prev := g.Dijkstra("A")
	for _, n := range g.Nodes() {
		fmt.Printf("A to %s: %2d via %s\n", n, dist[n], path(prev, n))
	}

	deps := NewGraph()
	for _, d := range [][2]string{{"shirt", "tie"}, {"tie", "jacket"}, {"trousers", "shoes"}, {"trousers", "belt"}, {"belt", "jacket"}, {"socks", "shoes"}, {"shirt", "belt"}} {
		deps.AddEdge(d[0], d[1], 1, true)
	}
	order, err := deps.TopoSort()
	fmt.Println("dress order:", order, err)
	deps.AddEdge("jacket", "shirt", 1, true)
	_, err = deps.TopoSort()
	fmt.Println("with cycle:", err)

	type wedge struct{ a, b, w int }
	edges := []wedge{{0, 1, 4}, {0, 2, 3}, {1, 2, 1}, {1, 3, 2}, {2, 3, 4}, {3, 4, 2}, {4, 5, 6}}
	sort.Slice(edges, func(i, j int) bool { return edges[i].w < edges[j].w })
	uf := NewUF(6)
	total := 0
	var used []string
	for _, e := range edges {
		if uf.Union(e.a, e.b) {
			total += e.w
			used = append(used, fmt.Sprintf("%d-%d", e.a, e.b))
		}
	}
	fmt.Println("MST weight", total, used)
}
