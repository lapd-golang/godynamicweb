package webapp

import (
	"fmt"

	"github.com/riotemergence/godynamicweb/multitenancy"
	"github.com/riotemergence/godynamicweb/server"
	"github.com/riotemergence/godynamicweb/x509"
)

type WebAppConfig struct {
	Name       *string                    `json:"name"`
	Version    *string                    `json:"version"`
	Connectors *server.ConnectorsConfig   `json:"connectors"`
	Tenants    *multitenancy.TenantConfig `json:"tenants"`
	X509       *x509.X509Config           `json:"tenants"`
}

func (c WebAppConfig) Validate() error {
	if c.Name == nil {
		return fmt.Errorf(TRACE + " WebAppConfig Name: required")
	}

	if c.Version == nil {
		return fmt.Errorf(TRACE + " WebAppConfig Name: required")
	}

	if c.Connectors == nil {
		return fmt.Errorf(TRACE + " WebAppConfig Connectors: required")
	}

	if err := c.Connectors.Validate(); err != nil {
		return err
	}

	if c.Tenants == nil {
		return fmt.Errorf(TRACE + " WebAppConfig Tenants: required")
	}

	if err := c.Tenants.Validate(); err != nil {
		return err
	}

	if c.X509 != nil {
		if err := c.X509.Validate(); err != nil {
			return err
		}
	}

	return nil
}
