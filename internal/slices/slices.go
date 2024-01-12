package slices

// Remove is a generic way to remove the first instance of 'value' from the given slice.
// If the item does not exist in the slice, you'll get back your input slice as-is. This
// does not mutate your input slice - it returns a new slice without the value.
func Remove[T comparable](slice []T, value T) []T {
	for i, e := range slice {
		if e == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// Map performs a functional Map operation where you convert a slice of type T to a slice of type U by
// running every element through a mapping function.
func Map[T any, U any](slice []T, mapper func(T) U) []U {
	if slice == nil {
		return nil
	}
	if mapper == nil {
		return nil
	}

	output := make([]U, len(slice))
	for i, item := range slice {
		output[i] = mapper(item)
	}
	return output
}

// Contains returns true if any of the slice elements pass an == test w/ the given 'value'.
func Contains[T comparable](slice []T, value T) bool {
	for _, sliceValue := range slice {
		if sliceValue == value {
			return true
		}
	}
	return false
}
