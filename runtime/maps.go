package runtime

func MapKeys[K comparable, V any](values map[K]V) []K {
	result := make([]K, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}

func MapValues[K comparable, V any](values map[K]V) []V {
	result := make([]V, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func MapContains[K comparable, V any](values map[K]V, key K) bool {
	_, ok := values[key]
	return ok
}

func MapClone[K comparable, V any](values map[K]V) map[K]V {
	if values == nil {
		return nil
	}
	result := make(map[K]V, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func MapClear[K comparable, V any](values map[K]V) {
	for key := range values {
		delete(values, key)
	}
}
