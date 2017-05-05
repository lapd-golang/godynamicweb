package server

import (
	"net"

	"fmt"

	"github.com/riotemergence/godynamicweb/util"
)

const (
	defaultHTTPPort  = 80
	defaultHTTPSPort = 443
)

// TCPPort a string representing a existing file path
type BindIpAddress string

// Set setter for ExistingFile
func (addr BindIpAddress) Validate() error {
	if string(addr) == "" {
		return fmt.Errorf(TRACE + " BindIpAddress: required")
	}
	ip := net.ParseIP(string(addr))
	if ip == nil {
		return fmt.Errorf(TRACE + " BindIpAddress: validity")
	}

	return nil
}

// TCPPort a string representing a existing file path
type TCPPort uint16

// Set setter for ExistingFile
func (tp TCPPort) Validate() error {
	if tp < 1 || tp > 65535 {
		return fmt.Errorf(TRACE + " TCPPort: rangeValidity")
	}
	return nil
}

type ConnectorConfig struct {
	BindAddress *BindIpAddress `json:"bindAddress"`
	Port        *TCPPort       `json:"port"`
	TLS         *bool          `json:"tls"`
}

func NewConnectorConfig(bindAddress string, port uint16, tls bool) *ConnectorConfig {
	return &ConnectorConfig{
		BindAddress: (*BindIpAddress)(&bindAddress),
		Port:        (*TCPPort)(&port),
		TLS:         (&tls),
	}
}

func (c ConnectorConfig) String() string {
	return util.ToJson(c)
}

func (c ConnectorConfig) Validate() error {
	if c.BindAddress == nil {
		return fmt.Errorf(TRACE + " ConnectorConfig BindAddress: required")
	}
	if err := c.BindAddress.Validate(); err != nil {
		return fmt.Errorf(TRACE+" ConnectorConfig BindAddress: %s", err)
	}
	if c.Port == nil {
		return fmt.Errorf(TRACE + " ConnectorConfig Port: required")
	}
	if err := c.Port.Validate(); err != nil {
		return fmt.Errorf(TRACE+" ConnectorConfig Port: %s", err)
	}
	if c.TLS == nil {
		return fmt.Errorf(TRACE + " ConnectorConfig TLS: required")
	}
	return nil
}

type ConnectorsConfig map[string]ConnectorConfig

func (c ConnectorsConfig) String() string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return util.ToJson(keys)
}

func (c ConnectorsConfig) Validate() error {
	for k, v := range c {
		if err := v.Validate(); err != nil {
			return fmt.Errorf(TRACE+" ConnectorsConfig \"%s\": %s", k, err)
		}
	}
	return nil
}

type ServerConfig struct {
	Connectors *ConnectorsConfig `json:"connectors"`
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Connectors: &ConnectorsConfig{},
	}
}

func (c ServerConfig) String() string {
	return util.ToJson(c)
}

func (sc ServerConfig) Validate() error {
	if sc.Connectors == nil {
		return fmt.Errorf(TRACE + " ServerConfig Connectors: required")
	}
	if err := sc.Connectors.Validate(); err != nil {
		return fmt.Errorf(TRACE+" ServerConfig Connectors: %s", err)
	}
	return nil
}
