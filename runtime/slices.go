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

func Index[T comparable](values []T, value T) int {
	for i, item := range values {
		if item == value {
			return i
		}
	}
	return -1
}

func Reverse[T any](values []T) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func Filter[T any](values []T, keep func(T) bool) []T {
	result := make([]T, 0, len(values))

	for _, value := range values {
		if keep(value) {
			result = append(result, value)
		}
	}

	return result
}

func Map[T any, R any](values []T, transform func(T) R) []R {
	result := make([]R, len(values))

	for i, value := range values {
		result[i] = transform(value)
	}

	return result
}

func Reduce[T any, R any](values []T, initial R, combine func(R, T) R) R {
	result := initial

	for _, value := range values {
		result = combine(result, value)
	}

	return result
}
