package runtime

import (
	"fmt"
	"strconv"
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

func Sprint(args ...any) string {
	return fmt.Sprint(args...)
}

func Sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

func Sprintln(args ...any) string {
	return fmt.Sprintln(args...)
}

func Itoa(value int) string {
	return strconv.Itoa(value)
}

func Atoi(value string) (int, error) {
	return strconv.Atoi(value)
}

func TrimSpace(value string) string {
	return strings.TrimSpace(value)
}

func Split(value, sep string) []string {
	return strings.Split(value, sep)
}

func Replace(value, old, new string, count int) string {
	return strings.Replace(value, old, new, count)
}

func ContainsString(value, part string) bool {
	return strings.Contains(value, part)
}

func HasPrefix(value, prefix string) bool {
	return strings.HasPrefix(value, prefix)
}

func HasSuffix(value, suffix string) bool {
	return strings.HasSuffix(value, suffix)
}
