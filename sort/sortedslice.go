package sort

import (
	"sort"
)

func Search(n int, fn func(i int) int) (lo int, hi int, found bool) {
	lo = sort.Search(n, func(i int) bool {
		return fn(i) <= 0
	})

	hi = sort.Search(n, func(i int) bool {
		return fn(i) < 0
	})
	found = (hi - lo) > 0
	return
}
