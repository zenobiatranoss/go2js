package javascript

import "testing"

func TestStdlibFuncName(t *testing.T) {
	tests := []struct {
		pkg  string
		name string
		want string
		ok   bool
	}{
		{"strings", "Contains", "go2jsStringsContains", true},
		{"strings", "TrimPrefix", "go2jsStringsTrimPrefix", true},
		{"strings", "TrimSuffix", "go2jsStringsTrimSuffix", true},
		{"strconv", "Unquote", "go2jsStrconvUnquote", true},
		{"strconv", "FormatFloat", "go2jsStrconvFormatFloat", true},
		{"math", "Exp", "Math.exp", true},
		{"math", "Sin", "Math.sin", true},
		{"math", "IsNaN", "Number.isNaN", true},
		{"sort", "Ints", "go2jsSortInts", true},
		{"unknown", "Unknown", "", false},
	}

	for _, test := range tests {
		got, ok := stdlibFuncName(test.pkg, test.name)
		if got != test.want || ok != test.ok {
			t.Fatalf("%s.%s = (%q, %v), want (%q, %v)", test.pkg, test.name, got, ok, test.want, test.ok)
		}
	}
}

func TestMultiReturnStdlibFuncs(t *testing.T) {
	for _, name := range []string{
		"Atoi",
		"ParseInt",
		"ParseFloat",
		"ParseBool",
	} {
		if !multiReturnStdlibFuncs["strconv."+name] {
			t.Fatalf("missing multi-return stdlib function %q", name)
		}
	}
}
