package runtime

import (
	"fmt"
	"strings"
)

func Format(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

func Print(args ...any) string {
	return fmt.Sprint(args...)
}

func Println(args ...any) string {
	return fmt.Sprintln(args...)
}

func Join(parts []string, sep string) string {
	return strings.Join(parts, sep)
}
