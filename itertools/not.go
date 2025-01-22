package itertools

func Not[T any](fn func(t T) bool) func(t T) bool {
	return func(t T) bool {
		return !fn(t)
	}
}
