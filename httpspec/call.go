package httpspec

type NamedHTTPOperationConfig struct {
	Name          string
	ParametersMap map[string]string   `json:"parametersMap"`
	Operation     HTTPOperationConfig `json:"operation"`
}

type ServiceClientEndpoint struct {
	Name       string
	Parameters map[string]interface{}     `json:"parameters"`
	Variables  map[string]string          `json:"variables"`
	Operations []NamedHTTPOperationConfig `json:"operations"`
}

type ServiceClientConfiguration struct {
	Endpoints map[string]ServiceClientEndpoint `json:"clientEndpoints"`
}
