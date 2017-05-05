package sort

import (
	"strings"

	math "github.com/riotemergence/godynamicweb/math"
)

func CompareStringArrays(a1 []string, a2 []string) int {
	l1, l2 := len(a1), len(a2)
	l := math.MinInt(l1, l2)
	for i := 0; i < l; i++ {
		s1, s2 := a1[i], a2[i]
		r := strings.Compare(s1, s2)
		if r != 0 {
			return r
		}
	}
	return l1 - l2
}
