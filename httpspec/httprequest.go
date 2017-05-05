package httpspec

import (
	"fmt"
	"net/http"
	"strings"
)

var knownHTTPMethods []string = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodOptions,
}

//HTTPRequestParams represents the possible dynamic parameters configuration of a HTTP Request call
type HTTPRequestParams struct {
	//Path A map of parameters to replace dynamic expressions built directly in the URL. Example: In the http://example.com/{folder}/{file} {folder} and {file} expressions can be replaced by values in the Path parameters map.
	Path map[string]string `json:"path,omitempty"`

	//Query A map of parameters to be added in the query string at the end of the HTTP URL in the format ?key1=value1&key2=value2 ...
	Query map[string]string `json:"query,omitempty"`

	//Headers A map of parameters to be added in the header of the HTTP call. Caution: The Headers are standartized
	Headers map[string]string `json:"headers,omitempty"`

	//Cookies A map of cookies to be added in the headers of the HTTP call.
	Cookies map[string]string `json:"cookies,omitempty"`
}

//HTTPRequestConfig a dynamically configured HTTP request
type HTTPRequestConfig struct {
	Method               string            `json:"method"`
	URL                  string            `json:"url"`
	Params               HTTPRequestParams `json:"params,omitempty"`
	EntityTransformation string            `json:"transformation,omitempty"`
}

func contains(array []string, test string) bool {
	for _, value := range array {
		if test == value {
			return true
		}
	}
	return false
}

func validateHTTPRequest(r HTTPRequestConfig) error {

	if r.Method == "" {
		return fmt.Errorf("HTTP method required")
	}
	if contains(knownHTTPMethods, r.Method) {
		return fmt.Errorf("HTTP method not supported: %s", r.Method)
	}
	if r.URL == "" {
		return fmt.Errorf("URL required")
	}
	if !(strings.HasPrefix(r.URL, "http://") || strings.HasPrefix(r.URL, "https://")) {
		return fmt.Errorf("Only HTTP URLs supported: %s", r.URL)
	}

	return nil
}
