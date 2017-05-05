package webapp

import (
	"fmt"
	"net/http"

	"github.com/riotemergence/godynamicweb/multitenancy"
	"github.com/riotemergence/godynamicweb/server"
	"github.com/riotemergence/godynamicweb/util"
)

func (webApp *WebApp) retrieveServerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, webApp.server.String())
}

func (webApp *WebApp) deleteServerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Shutting down connector")
	webApp.server.Stop()
}

func (webApp *WebApp) listServerConnectorsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, webApp.server.Config.Connectors)
}

func (webApp *WebApp) createOrReplaceServerConnectorHandler(w http.ResponseWriter, r *http.Request) {
	var c server.ConnectorConfig
	util.Put(w, r, "connectorName",
		func(connectorName string) bool {
			_, found := (*webApp.server.Config.Connectors)[connectorName]
			return found
		},
		func(connectorName string) error {
			fmt.Println("Create")
			if err := webApp.CreateServerConnector(connectorName, c); err != nil {
				return err
			}
			return nil
		},
		func(connectorName string) error {
			fmt.Println("Update")
			return fmt.Errorf("Update not Permitted")
		},
		&c,
	)
}

func (webApp *WebApp) retrieveServerConnectorHandler(w http.ResponseWriter, r *http.Request) {
	util.Get(w, r, "connectorName", func(connectorName string) (fmt.Stringer, bool) {
		r, ok := (*webApp.server.Config.Connectors)[connectorName]
		return r, ok
	})
}

func (webApp *WebApp) deleteServerConnectorHandler(w http.ResponseWriter, r *http.Request) {
	err := util.Delete(w, r, "connectorName",
		func(connectorName string) bool {
			_, ok := (*webApp.server.Config.Connectors)[connectorName]
			return ok
		},
		func(connectorName string) error {
			return webApp.DeleteServerConnector(connectorName)
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
	}
}

func (webApp *WebApp) listTenantsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, webApp.multiTenancySupport.Config.Tenants)
}

func (webApp *WebApp) createOrReplaceTenantHandler(w http.ResponseWriter, r *http.Request) {
	var c multitenancy.TenantConfig
	util.Put(w, r, "tenantID",
		func(tenantID string) bool {
			_, found := webApp.multiTenancySupport.Config.Tenants[tenantID]
			return found
		},
		func(tenantID string) error {
			fmt.Println("Create")
			if err := webApp.CreateTenant(tenantID, c); err != nil {
				return err
			}
			return nil
		},
		func(tenantID string) error {
			fmt.Println("Update")
			return fmt.Errorf("Update not Permitted")
		},
		&c,
	)
}

func (webApp *WebApp) retrieveTenantHandler(w http.ResponseWriter, r *http.Request) {
	util.Get(w, r, "tenantID", func(tenantID string) (fmt.Stringer, bool) {
		r, ok := webApp.multiTenancySupport.Config.Tenants[tenantID]
		return r, ok
	})
}

func (webApp *WebApp) deleteTenantHandler(w http.ResponseWriter, r *http.Request) {
	err := util.Delete(w, r, "tenantID",
		func(tenantID string) bool {
			_, ok := webApp.multiTenancySupport.Config.Tenants[tenantID]
			return ok
		},
		func(tenantID string) error {
			return webApp.DeleteTenant(tenantID)
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
	}
}
