package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

func ToJson(i interface{}) string {
	jsonBytes, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func KeySet(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Get(w http.ResponseWriter, r *http.Request, pathParameterName string, extractParameterFn func(string) (fmt.Stringer, bool)) {
	pathParameters := mux.Vars(r)
	pathParameterValue := pathParameters[pathParameterName]
	if s, ok := extractParameterFn(pathParameterValue); ok {
		fmt.Fprint(w, s.String())
		return
	}
	http.NotFound(w, r)
}

func Delete(w http.ResponseWriter, r *http.Request, pathParameterName string, existsParameterFn func(string) bool, execFn func(string) error) error {
	pathParameters := mux.Vars(r)
	pathParameterValue := pathParameters[pathParameterName]
	if ok := existsParameterFn(pathParameterValue); !ok {
		http.NotFound(w, r)
		return nil
	}
	return execFn(pathParameterValue)
}

func Put(w http.ResponseWriter, r *http.Request, pathParameterName string, existsParameterFn func(string) bool, createFn func(string) error, updateFn func(string) error, bodyParamPtr interface{}) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(bodyParamPtr); err != nil {
		http.Error(w, "Invalid JSON Body", http.StatusConflict)
		return err
	}
	pathParameters := mux.Vars(r)
	pathParameterValue := pathParameters[pathParameterName]
	if ok := existsParameterFn(pathParameterValue); !ok {
		if err := createFn(pathParameterValue); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return err
		}
		w.WriteHeader(http.StatusCreated)
		return nil

	}

	if err := updateFn(pathParameterValue); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return err
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func GetMandatoryUniqueParam(queryValues url.Values, queryParamName string) (string, error) {
	if len(queryValues[queryParamName]) == 0 {
		return "", fmt.Errorf("%s parameter is mandatory", queryParamName)
	}

	if len(queryValues[queryParamName]) != 1 {
		return "", fmt.Errorf("%s parameter must be unique", queryParamName)
	}
	return queryValues[queryParamName][0], nil
}
