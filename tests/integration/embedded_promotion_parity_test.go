package integration_test

import "testing"

func TestEmbeddedPointerMethodPromotion(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type inner struct{ msg string }

func (i *inner) Error() string { return "inner:" + i.msg }

func (i *inner) Describe() string { return "described " + i.msg }

type outer struct {
	*inner
	extra int
}

func main() {
	o := &outer{inner: &inner{msg: "hi"}, extra: 3}

	fmt.Println(o)
	fmt.Println(o.Error())
	fmt.Println(o.Describe())
	fmt.Printf("%+v %v\n", o, o)

	var err error = o
	fmt.Println(err)
	fmt.Println(err != nil)
}
`)
}

func TestEmbeddedValueMethodPromotionOverridesOuterMethod(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type base struct{ id int }

func (b base) Name() string  { return "base" }
func (b base) Extra() string { return "extra" }

type derived struct {
	base
	label string
}

func (d derived) Name() string { return "derived+" + d.base.Name() }

func main() {
	d := derived{base: base{id: 7}, label: "x"}

	fmt.Println(d.Name())
	fmt.Println(d.Extra())
	fmt.Println(d.id, d.label)

	ptr := &d
	fmt.Println(ptr.Name())
	fmt.Println(ptr.Extra())
}
`)
}
