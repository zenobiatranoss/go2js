package integration_test

import "testing"

func TestHeapInterfaceAndTableTypeMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"container/heap"
	"fmt"
	"hash/crc32"
	"hash/crc64"
)

type pq []int

func (p pq) Len() int            { return len(p) }
func (p pq) Less(i, j int) bool  { return p[i] < p[j] }
func (p pq) Swap(i, j int)       { p[i], p[j] = p[j], p[i] }
func (p *pq) Push(x interface{}) { *p = append(*p, x.(int)) }
func (p *pq) Pop() interface{} {
	old := *p
	n := len(old)
	x := old[n-1]
	*p = old[:n-1]
	return x
}

func main() {
	var h heap.Interface = &pq{}
	heap.Push(h, 3)
	heap.Push(h, 1)
	heap.Push(h, 2)
	fmt.Println(h.Len(), heap.Pop(h), heap.Pop(h), heap.Pop(h))

	var t *crc32.Table = crc32.MakeTable(crc32.IEEE)
	fmt.Printf("%08x\n", crc32.Checksum([]byte("hello"), t))
	fmt.Printf("%08x\n", crc32.Checksum([]byte("hello"), crc32.IEEETable))

	var t64 *crc64.Table = crc64.MakeTable(crc64.ISO)
	fmt.Printf("%016x\n", crc64.Checksum([]byte("hello"), t64))
}
`)
}
