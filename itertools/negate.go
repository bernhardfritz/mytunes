package itertools

func Negate[T any](fn func(t T) bool) func(t T) bool {
	return func(t T) bool {
		return !fn(t)
	}
}
