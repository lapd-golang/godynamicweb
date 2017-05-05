package sort

import (
	"net/url"
	"strings"
)

func CompareURLs(u1, u2 *url.URL) int {
	comparisonResult := strings.Compare(u1.Scheme, u2.Scheme)
	if comparisonResult != 0 {
		return comparisonResult
	}
	comparisonResult = strings.Compare(u1.Host, u2.Host)
	if comparisonResult != 0 {
		return comparisonResult
	}

	splittedPaths1, splittedPaths2 := strings.Split(u1.Path, "/"), strings.Split(u2.Path, "/")
	comparisonResult = CompareStringArrays(splittedPaths1[1:], splittedPaths2[1:])
	if comparisonResult != 0 {
		return comparisonResult
	}

	//TODO Query Parameters comparison

	return comparisonResult
}
