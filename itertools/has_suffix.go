package itertools

import "strings"

func HasSuffix(suffix string) func(s string) bool {
	return func(s string) bool {
		return strings.HasSuffix(s, suffix)
	}
}
