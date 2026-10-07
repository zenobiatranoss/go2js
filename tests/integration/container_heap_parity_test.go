package integration_test

import "testing"

// The heap contracts ask the heap for its length, its comparisons and its
// swaps, and then do the walking between the slots itself: a deferred push
// bubbles the value that was just appended up, a pop walks the new root down,
// and a fix walks whichever way the changed value has to go. The heap in this
// test keeps track of the slot every item stands in, so the walk has to honor
// the swaps the heap itself performs.
func TestContainerHeapRunsTheDocExample(t *testing.T) {
	runParityTest(t, `package main

import (
	"container/heap"
	"fmt"
)

type Item struct {
	value    string
	priority int
	index    int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func update(pq *PriorityQueue, item *Item, value string, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

func main() {
	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &Item{value: "banana", priority: 3})
	heap.Push(pq, &Item{value: "apple", priority: 4})
	heap.Push(pq, &Item{value: "pear", priority: 5})
	item := &Item{value: "orange", priority: 2}
	heap.Push(pq, item)
	update(pq, item, "grape", 1)
	for pq.Len() > 0 {
		item := heap.Pop(pq).(*Item)
		fmt.Printf("%.2d:%s ", item.priority, item.value)
	}
	fmt.Println()

	c := &PriorityQueue{}
	heap.Push(c, &Item{value: "a", priority: 4})
	heap.Push(c, &Item{value: "b", priority: 2})
	heap.Push(c, &Item{value: "c", priority: 5})
	top := heap.Pop(c).(*Item)
	fmt.Println("top", top.value)
	removed := heap.Remove(c, 1).(*Item)
	fmt.Println("removed", removed.value)
	for c.Len() > 0 {
		it := heap.Pop(c).(*Item)
		fmt.Printf("%.2d:%s ", it.priority, it.value)
	}
	fmt.Println()
}
`)
}

// A heap whose type is a plain slice works the same way, and is the shape most
// of the documentation writes it in.
func TestContainerHeapOnAPlainSlice(t *testing.T) {
	runParityTest(t, `package main

import (
	"container/heap"
	"fmt"
)

type IntHeap []int

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x any) {
	*h = append(*h, x.(int))
}

func (h *IntHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func main() {
	h := &IntHeap{2, 1, 5}
	heap.Init(h)
	heap.Push(h, 3)
	fmt.Printf("min: %d\n", (*h)[0])
	for h.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(h))
	}
	fmt.Println()
}
`)
}
