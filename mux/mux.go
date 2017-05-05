package mux

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	math "github.com/riotemergence/godynamicweb/math"
	sort "github.com/riotemergence/godynamicweb/sort"
)

const TRACE = "github.com/riotemergence/godynamicweb/mux"

type PathParts []string

func NewPathParts(path string) *PathParts {
	pathParts := PathParts{}
	if path == "" {
		return &pathParts
	}

	for strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	pathParts = strings.Split(path, "/")
	return &pathParts
}

func (p *PathParts) String() string {
	var buffer bytes.Buffer
	for _, v := range *p {
		buffer.WriteString("/")
		buffer.WriteString(v)
	}
	return buffer.String()
}

type MuxKey struct {
	Connector string
	Scheme    string
	Host      string
	Path      PathParts
	Method    string
	//TODO QueryParams map[string][]string

}

func (k *MuxKey) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	buffer.WriteString(k.Connector)
	buffer.WriteString("] ")
	buffer.WriteString(k.Method)
	buffer.WriteString(" ")
	buffer.WriteString(k.Scheme)
	buffer.WriteString("://")
	buffer.WriteString(k.Host)
	for _, p := range k.Path {
		buffer.WriteString("/")
		buffer.WriteString(p)
	}
	return buffer.String()
}

type MuxEntry struct {
	Key   MuxKey
	Value interface{}
}

type MuxCatalog []MuxEntry

func NewMuxCatalog() *MuxCatalog {
	muxCatalog := make(MuxCatalog, 0)
	return &muxCatalog
}

func (mc *MuxCatalog) Add(connector, scheme, host, path, method string, value interface{}) error {
	//FIXME Some bug
	muxEntry := MuxEntry{
		Key: MuxKey{
			Connector: connector,
			Scheme:    scheme,
			Host:      host,
			Path:      *NewPathParts(path),
			Method:    method,
		},
		Value: value,
	}

	mcLen := len(*mc)
	insertionPointIndex, _, found := sort.Search(mcLen,
		func(compareIndex int) int {
			return CompareMuxEntry(muxEntry, (*mc)[compareIndex])
		},
	)

	if found {
		conflictingEntry := (*mc)[insertionPointIndex]
		return fmt.Errorf(TRACE+" MuxCatalog Add: mustNotConflictWithExistingEntry \"%s\"", conflictingEntry)
	}

	*mc = append(*mc, MuxEntry{})
	copy((*mc)[insertionPointIndex+1:], (*mc)[insertionPointIndex:])
	(*mc)[insertionPointIndex] = muxEntry
	return nil
}

func (mc *MuxCatalog) Remove(index int) {
	*mc = append((*mc)[:index], (*mc)[index+1:]...)
	//return mc
}

func (mc *MuxCatalog) RemoveAll(removeWhen func(muxEntry MuxEntry) bool) {
	temp := (*mc)[:0]
	for _, v := range *mc {
		if removeWhen(v) {
			temp = append(temp, v)
		}
	}
	*mc = temp
}

func (mc *MuxCatalog) GetWithRequest(connectorName string, r *http.Request) (*MuxEntry, bool) {
	mcLen := len(*mc)
	lo, _, found := sort.Search(mcLen, func(compareIndex int) int {
		return CompareRequestVsMuxEntry(connectorName, r, (*mc)[compareIndex])
	})

	if !found {
		return nil, false
	}
	entry := &MuxEntry{}
	*entry = (*mc)[lo]
	return entry, true

}

func comparePaths(path1Parts, path2Parts []string, path1IsDynamic bool) int {
	path1PartsLength, path2PartsLength := len(path1Parts), len(path2Parts)
	if path1PartsLength == 1 && len(path1Parts[0]) == 0 {
		path1PartsLength = 0
		path1Parts = []string{}
	}

	if path2PartsLength == 1 && len(path2Parts[0]) == 0 {
		path2PartsLength = 0
		path2Parts = []string{}
	}

	pathPartsCommonLength := math.MinInt(path1PartsLength, path2PartsLength)
	path1PartsLastIndex := path1PartsLength - 1
	path2PartsLastIndex := path2PartsLength - 1

	for pathPartIndex := 0; pathPartIndex < pathPartsCommonLength; pathPartIndex++ {
		path1Part, path2Part := path1Parts[pathPartIndex], path2Parts[pathPartIndex]

		if path1IsDynamic && pathPartIndex == path1PartsLastIndex && path1Part == "*" {
			return 0
		}

		if pathPartIndex == path2PartsLastIndex && path2Part == "*" {
			return 0
		}

		path1PartIsDynamic := strings.HasPrefix(path1Part, "{") && strings.HasSuffix(path1Part, "}")
		path2PartIsDynamic := strings.HasPrefix(path2Part, "{") && strings.HasSuffix(path2Part, "}")
		if (path1IsDynamic && path1PartIsDynamic) || path2PartIsDynamic {
			continue
		}

		comparisonResult := strings.Compare(path1Part, path2Part)
		if comparisonResult != 0 {
			return comparisonResult
		}
	}

	comparisonResult := path1PartsLength - path2PartsLength
	if comparisonResult != 0 {
		return comparisonResult
	}

	return 0
}

func CompareMuxEntry(entry1, entry2 MuxEntry) int {
	key1, key2 := entry1.Key, entry2.Key

	comparisonResult := strings.Compare(key1.Connector, key2.Connector)
	if comparisonResult != 0 {
		return comparisonResult
	}

	comparisonResult = strings.Compare(key1.Scheme, key2.Scheme)
	if comparisonResult != 0 {
		return comparisonResult
	}

	comparisonResult = strings.Compare(key1.Host, key2.Host)
	if comparisonResult != 0 {
		return comparisonResult
	}

	comparisonResult = comparePaths(key1.Path, key2.Path, true)
	if comparisonResult != 0 {
		return comparisonResult
	}

	comparisonResult = strings.Compare(key1.Method, key2.Method)
	if comparisonResult != 0 {
		return comparisonResult
	}

	return 0
}

func CompareRequestVsMuxEntry(reqConnectorName string, req *http.Request, muxEntry MuxEntry) int {
	muxKey := muxEntry.Key
	comparisonResult := strings.Compare(reqConnectorName, muxKey.Connector)
	if comparisonResult != 0 {
		return comparisonResult
	}

	//TODO Get Scheme
	var reqScheme string
	if req.TLS != nil {
		reqScheme = "https"
	} else {
		reqScheme = "http"
	}
	if forwardedProto := req.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		reqScheme = forwardedProto
	}
	comparisonResult = strings.Compare(reqScheme, muxKey.Scheme)
	if comparisonResult != 0 {
		return comparisonResult
	}

	host := req.Host
	if forwardedHost := req.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}
	comparisonResult = strings.Compare(host, muxKey.Host)
	if comparisonResult != 0 {
		return comparisonResult
	}
	comparisonResult = comparePaths(*NewPathParts(req.RequestURI), muxKey.Path, false)
	if comparisonResult != 0 {
		return comparisonResult
	}

	comparisonResult = strings.Compare(req.Method, muxKey.Method)
	if comparisonResult != 0 {
		return comparisonResult
	}

	return 0
}
