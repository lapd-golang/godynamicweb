package httpspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type HTTPOperationConfig struct {
	Request  HTTPRequestConfig  `json:"request"`
	Response HTTPResponseConfig `json:"response"`
}

func NewHTTPOperationConfigFromJSON(specURL string) (*HTTPOperationConfig, error) {

	var jsonData io.Reader
	if strings.HasPrefix(specURL, "http://") || strings.HasPrefix(specURL, "https://") {
		r, err := http.Get(specURL)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		jsonData = r.Body
	} else {
		f, err := os.Open(specURL)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		jsonData = f
	}
	jsonDecoder := json.NewDecoder(jsonData)
	spec := &HTTPOperationConfig{}
	if err := jsonDecoder.Decode(spec); err != nil {
		return nil, fmt.Errorf("client JSON content invalid: %s", err.Error())
	}

	validateHTTPRequest(spec.Request)

	return spec, nil
}

func (spec HTTPOperationConfig) NewRequest(vars map[string]string, bodyEntity *interface{}) (*http.Request, error) {
	//TODO Body

	u, err := fillURLParameters(spec, vars)
	if err != nil {
		return nil, err
	}

	bodyEntityReader := io.Reader(nil)

	if bodyEntity != nil {
		bodyEntityBytes, err := json.Marshal(bodyEntity)
		if err != nil {
			return nil, err
		}
		bodyEntityReader = bytes.NewBuffer(bodyEntityBytes)
	}

	r, err := http.NewRequest(strings.ToUpper(spec.Request.Method), u, bodyEntityReader)
	if err != nil {
		return nil, err
	}

	for headerName, headerValue := range spec.Request.Params.Headers {
		r.Header.Add(headerName, headerValue)
	}

	for cookieName, cookieValue := range spec.Request.Params.Cookies {
		cookie := http.Cookie{}
		cookie.Name = cookieName
		cookie.Value = cookieValue
		r.AddCookie(&cookie)
	}

	return r, nil
}

func (spec HTTPOperationConfig) String() string {
	jsonBytes, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func fillURLParameters(spec HTTPOperationConfig, vars map[string]string) (string, error) {
	u := spec.Request.URL
	for k, v := range spec.Request.Params.Path {
		q := regexp.QuoteMeta("{" + k + "}")
		r := regexp.MustCompile("(/)(" + q + ")(/|$)")
		s := varSubstitution(v, vars)
		u = r.ReplaceAllString(u, "${1}"+url.QueryEscape(s)+"${3}")
	}

	pu, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	if !pu.IsAbs() {
		return "", fmt.Errorf("invalid value for spec.Request.URL: %s must be an absolute URL", u)
	}

	q := pu.Query()
	for k, v := range spec.Request.Params.Query {
		q.Add(k, url.QueryEscape(varSubstitution(v, vars)))
	}
	pu.RawQuery = q.Encode()

	return pu.String(), nil
}

func varSubstitution(expression string, vars map[string]string) string {
	u := expression
	for k, v := range vars {
		q := regexp.QuoteMeta("{" + k + "}")
		r := regexp.MustCompile("(" + q + ")")
		u = r.ReplaceAllString(u, v)
	}
	return u
}
