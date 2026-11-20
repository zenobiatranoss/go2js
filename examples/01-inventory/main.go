// Inventory: structs, interfaces, embedding, sorting and formatted reports.
package main

import (
	"fmt"
	"sort"
	"strings"
)

type Item interface {
	SKU() string
	Name() string
	Price() float64
	Describe() string
}

type Base struct {
	sku   string
	name  string
	price float64
}

func (b Base) SKU() string      { return b.sku }
func (b Base) Name() string     { return b.name }
func (b Base) Price() float64   { return b.price }
func (b Base) Describe() string { return fmt.Sprintf("%s (%s)", b.name, b.sku) }

type Book struct {
	Base
	Author string
	Pages  int
}

func (b Book) Describe() string {
	return fmt.Sprintf("%s by %s, %d pages", b.Name(), b.Author, b.Pages)
}

type Gadget struct {
	Base
	Watts int
}

func (g Gadget) Describe() string {
	return fmt.Sprintf("%s drawing %dW", g.Name(), g.Watts)
}

type Grocery struct {
	Base
	Perishable bool
}

type Stock struct {
	item Item
	qty  int
}

type Inventory struct {
	stock map[string]*Stock
}

func NewInventory() *Inventory { return &Inventory{stock: map[string]*Stock{}} }

func (inv *Inventory) Add(item Item, qty int) {
	if s, ok := inv.stock[item.SKU()]; ok {
		s.qty += qty
		return
	}
	inv.stock[item.SKU()] = &Stock{item, qty}
}

func (inv *Inventory) Remove(sku string, qty int) error {
	s, ok := inv.stock[sku]
	if !ok {
		return fmt.Errorf("unknown sku %q", sku)
	}
	if s.qty < qty {
		return fmt.Errorf("only %d of %s left, wanted %d", s.qty, sku, qty)
	}
	s.qty -= qty
	if s.qty == 0 {
		delete(inv.stock, sku)
	}
	return nil
}

func (inv *Inventory) Sorted(less func(a, b *Stock) bool) []*Stock {
	list := make([]*Stock, 0, len(inv.stock))
	for _, s := range inv.stock {
		list = append(list, s)
	}
	sort.Slice(list, func(i, j int) bool { return less(list[i], list[j]) })
	return list
}

func (inv *Inventory) Total() float64 {
	total := 0.0
	for _, s := range inv.stock {
		total += s.item.Price() * float64(s.qty)
	}
	return total
}

func report(title string, list []*Stock) {
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", 58))
	for _, s := range list {
		fmt.Printf("%-8s %-34s %4d x %7.2f\n", s.item.SKU(), s.item.Describe(), s.qty, s.item.Price())
	}
	fmt.Println()
}

func main() {
	inv := NewInventory()
	inv.Add(Book{Base{"B-100", "Go in Practice", 39.9}, "A. Smith", 320}, 4)
	inv.Add(Book{Base{"B-200", "Types and Programming", 54.5}, "B. Jones", 623}, 2)
	inv.Add(Gadget{Base{"G-001", "Desk lamp", 24.0}, 12}, 10)
	inv.Add(Gadget{Base{"G-002", "Kettle", 31.25}, 1800}, 3)
	inv.Add(Grocery{Base{"F-900", "Coffee beans", 12.75}, false}, 25)
	inv.Add(Grocery{Base{"F-901", "Fresh milk", 1.2}, true}, 40)
	inv.Add(Gadget{Base{"G-001", "Desk lamp", 24.0}, 12}, 5)

	report("By SKU", inv.Sorted(func(a, b *Stock) bool { return a.item.SKU() < b.item.SKU() }))
	report("By value, highest first", inv.Sorted(func(a, b *Stock) bool {
		return a.item.Price()*float64(a.qty) > b.item.Price()*float64(b.qty)
	}))

	fmt.Printf("Total stock value: %.2f\n", inv.Total())
	for _, step := range []struct {
		sku string
		qty int
	}{{"G-001", 15}, {"F-901", 41}, {"X-000", 1}, {"B-200", 2}} {
		if err := inv.Remove(step.sku, step.qty); err != nil {
			fmt.Println("error:", err)
		} else {
			fmt.Printf("removed %d of %s\n", step.qty, step.sku)
		}
	}
	fmt.Printf("Items left: %d, value %.2f\n", len(inv.stock), inv.Total())

	kinds := map[string]int{}
	for _, s := range inv.stock {
		switch s.item.(type) {
		case Book:
			kinds["book"]++
		case Gadget:
			kinds["gadget"]++
		default:
			kinds["other"]++
		}
	}
	fmt.Println(kinds)
}
