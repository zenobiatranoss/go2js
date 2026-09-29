package integration_test

import "testing"

func TestEmptyStatementAndLabelsParity(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	total := 0
	for i := 0; i < 4; i++ {
		if i == 2 {
			continue
		}
		total += i
	}

	;

	switch {
	case false:
		;
	default:
		fmt.Println("default")
	}

	if total > 100 {
		goto end
	}

	_ = total

end:
	;
	fmt.Println(total)
}
`)
}
