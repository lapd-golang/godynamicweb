package multitenancy

import (
	"fmt"
	"net/url"
	"os"

	"github.com/riotemergence/godynamicweb/util"
	"github.com/riotemergence/godynamicweb/x509"
)

// AbsoluteHttpUrl a URL that checks if it is absolute
type AbsoluteHttpUrl string

// Set setter for AbsoluteURL
func (u AbsoluteHttpUrl) Validate() error {
	uAsUrl, err := url.Parse(string(u))
	if err != nil {
		return fmt.Errorf(TRACE + " AbsoluteHttpUrl: mustBeValidUrl")
	}
	if !uAsUrl.IsAbs() {
		return fmt.Errorf(TRACE + " AbsoluteHttpUrl: mustBeAbsoluteUrl")
	}
	if uAsUrl.Scheme != "http" && uAsUrl.Scheme != "https" {
		return fmt.Errorf(TRACE + " AbsoluteHttpUrl: mustBeHttpUrl")
	}
	return nil
}

type ServerEndpointConfig struct {
	Url       *AbsoluteHttpUrl `json:"url"`
	Connector *string          `json:"connector"`
}

func (sec ServerEndpointConfig) Validate() error {
	if sec.Url == nil {
		return fmt.Errorf(TRACE + " ServerEndpointConfig Url: required")
	}
	if err := sec.Url.Validate(); err != nil {
		return fmt.Errorf(TRACE+" ServerEndpointConfig Url: %s", err)
	}
	if sec.Connector == nil {
		return fmt.Errorf(TRACE + " ServerEndpointConfig Connector: required")
	}
	return nil
}

type ServerEndpointsConfig map[string]ServerEndpointConfig

func (c ServerEndpointsConfig) Validate() error {
	for k, v := range c {
		if err := v.Validate(); err != nil {
			return fmt.Errorf(TRACE+" ServerEndpointsConfig \"%s\": %s", k, err)
		}
	}
	return nil
}

type ReverseProxyEndpointConfig struct {
	Url       *AbsoluteHttpUrl `json:"url"`
	Connector *string          `json:"connector"`
	Methods   *[]string        `json:"methods"`
	TargetUrl *AbsoluteHttpUrl `json:"targetUrl"`
}

func (rpc ReverseProxyEndpointConfig) Validate() error {
	if rpc.Url == nil {
		return fmt.Errorf(TRACE + " ReverseProxyConfig Url: required")
	}
	if err := rpc.Url.Validate(); err != nil {
		return fmt.Errorf(TRACE+" ReverseProxyConfig Url: %s", err)
	}
	if rpc.Connector == nil {
		return fmt.Errorf(TRACE + " ReverseProxyConfig Connector: required")
	}
	if rpc.Methods == nil || len(*rpc.Methods) == 0 {
		return fmt.Errorf(TRACE + " ReverseProxyConfig Methods: required")
	}
	if rpc.TargetUrl == nil {
		return fmt.Errorf(TRACE + " ReverseProxyConfig TargetUrl: required")
	}
	if err := rpc.TargetUrl.Validate(); err != nil {
		return fmt.Errorf(TRACE+" ReverseProxyConfig TargetUrl: %s", err)
	}

	return nil
}

type ReverseProxyEndpointsConfig []ReverseProxyEndpointConfig

func (c ReverseProxyEndpointsConfig) Validate() error {
	for k, v := range c {
		if err := v.Validate(); err != nil {
			return fmt.Errorf(TRACE+" ReverseProxyEndpointsConfig \"%s\": %s", k, err)
		}
	}
	return nil
}

// ExistingFile a string representing a existing file path
type ExistingDir string

// Set setter for ExistingFile
func (ed ExistingDir) Validate() error {
	fi, err := os.Stat(string(ed))
	if os.IsNotExist(err) {
		return fmt.Errorf("\"%s\" file not found", ed)
	}
	if !fi.IsDir() {
		return fmt.Errorf("\"%s\" must be dir", ed)
	}
	return nil
}

type FileServerEndpointConfig struct {
	Url        *AbsoluteHttpUrl `json:"url"`
	Connector  *string          `json:"connector"`
	RootFs     *ExistingDir     `json:"rootFs"`
	DirListing *bool            `json:"dirListing"`
}

func (fsec FileServerEndpointConfig) Validate() error {
	if fsec.Url == nil {
		return fmt.Errorf(TRACE + " FileServerEndpointConfig Url: required")
	}
	if err := fsec.Url.Validate(); err != nil {
		return fmt.Errorf(TRACE+" FileServerEndpointConfig Url: %s", err)
	}
	if fsec.Connector == nil {
		return fmt.Errorf(TRACE + " FileServerEndpointConfig Connector: required")
	}
	if fsec.RootFs == nil {
		return fmt.Errorf(TRACE + " FileServerEndpointConfig RootFs: required")
	}
	if err := fsec.RootFs.Validate(); err != nil {
		return fmt.Errorf(TRACE+" FileServerEndpointConfig RootFs: %s", err)
	}

	return nil
}

type FileServerEndpointsConfig []FileServerEndpointConfig

func (c FileServerEndpointsConfig) Validate() error {
	for k, v := range c {
		if err := v.Validate(); err != nil {
			return fmt.Errorf(TRACE+" FileServerEndpointsConfig \"%s\": %s", k, err)
		}
	}
	return nil
}

type TenantConfig struct {
	Name                  *string                      `json:"name"`
	X509                  *x509.X509Config             `json:"x509"`
	ServerEndpoints       *ServerEndpointsConfig       `json:"serverEndpoints"`
	ReverseProxyEndpoints *ReverseProxyEndpointsConfig `json:"reverseProxyEndpoints"`
	FileServerEndpoints   *FileServerEndpointsConfig   `json:"fileServerEndpoints"`
}

func (c TenantConfig) String() string {
	return util.ToJson(c)
}

func (c TenantConfig) Validate() error {
	if c.Name == nil {
		return fmt.Errorf(TRACE + " TenantConfig Name: required")
	}

	if c.X509 != nil {
		if err := c.X509.Validate(); err != nil {
			return err
		}
	}

	if c.ServerEndpoints == nil {
		return fmt.Errorf(TRACE + " TenantConfig ServerEndpoints: required")
	}
	if err := c.ServerEndpoints.Validate(); err != nil {
		return err
	}
	if c.ReverseProxyEndpoints != nil {
		if err := c.ReverseProxyEndpoints.Validate(); err != nil {
			return err
		}
	}
	if c.FileServerEndpoints != nil {
		if err := c.FileServerEndpoints.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type TenantsConfig map[string]TenantConfig

func (c TenantsConfig) String() string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return util.ToJson(keys)
}

type MultiTenancyConfig struct {
	Tenants TenantsConfig `json:"tenants"`
}
