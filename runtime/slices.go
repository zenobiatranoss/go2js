package runtime

func Append[T any](values []T, items ...T) []T {
	return append(values, items...)
}

func Copy[T any](dst, src []T) int {
	return copy(dst, src)
}

func Clear[T any](values []T) {
	var zero T
	for i := range values {
		values[i] = zero
	}
}

func Clone[T any](values []T) []T {
	if values == nil {
		return nil
	}
	result := make([]T, len(values))
	copy(result, values)
	return result
}

func Contains[T comparable](values []T, value T) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
