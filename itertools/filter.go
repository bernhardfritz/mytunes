package itertools

import "iter"

func Filter[T any](fn func(T) bool, seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for t := range seq {
			if fn(t) {
				if !yield(t) {
					return
				}
			}
		}
	}
}
