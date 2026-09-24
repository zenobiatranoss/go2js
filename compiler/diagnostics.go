package compiler

import "fmt"

type Diagnostic struct {
	Message string
}

type Diagnostics struct {
	Items []Diagnostic
}

func (d *Diagnostics) Add(message string) {
	d.Items = append(d.Items, Diagnostic{Message: message})
}

func (d *Diagnostics) Empty() bool {
	return len(d.Items) == 0
}

func (d *Diagnostics) Error() error {
	if d.Empty() {
		return nil
	}

	return fmt.Errorf("%s", d.Items[0].Message)
}
