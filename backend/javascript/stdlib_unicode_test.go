package javascript

import "testing"

func TestUnicodeStdlibMappings(t *testing.T) {
	cases := []struct {
		pkg  string
		name string
		want string
	}{
		{"unicode", "IsLetter", "go2jsUnicodeIsLetter"},
		{"unicode", "IsDigit", "go2jsUnicodeIsDigit"},
		{"unicode", "IsSpace", "go2jsUnicodeIsSpace"},
		{"unicode", "IsNumber", "go2jsUnicodeIsNumber"},
		{"unicode", "IsUpper", "go2jsUnicodeIsUpper"},
		{"unicode", "ToLower", "go2jsUnicodeToLower"},
		{"unicode", "ToUpper", "go2jsUnicodeToUpper"},
		{"utf8", "RuneCount", "go2jsUTF8RuneCount"},
		{"utf8", "RuneCountInString", "go2jsUTF8RuneCountInString"},
		{"utf8", "RuneLen", "go2jsUTF8RuneLen"},
		{"utf8", "RuneStart", "go2jsUTF8RuneStart"},
		{"utf8", "Valid", "go2jsUTF8Valid"},
		{"utf8", "ValidRune", "go2jsUTF8ValidRune"},
		{"utf8", "ValidString", "go2jsUTF8ValidString"},
	}

	for _, test := range cases {
		got, ok := stdlibFuncName(test.pkg, test.name)
		if !ok {
			t.Fatalf("%s.%s was not mapped", test.pkg, test.name)
		}

		if got != test.want {
			t.Fatalf("%s.%s = %q, want %q", test.pkg, test.name, got, test.want)
		}
	}
}
