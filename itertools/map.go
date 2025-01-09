package itertools

import "iter"

func Map[T any, U any](fn func(T) U, seq iter.Seq[T]) iter.Seq[U] {
	return func(yield func(U) bool) {
		for t := range seq {
			u := fn(t)
			if !yield(u) {
				return
			}
		}
	}
}
