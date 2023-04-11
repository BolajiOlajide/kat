package utils

func SliceIncludes[T string](s []T, e T) bool {
	for _, item := range s {
		if item == e {
			return true
		}
	}
	return false
}
