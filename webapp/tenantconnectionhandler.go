package webapp

import (
	"net/http"
	"net/http/httputil"

	"fmt"

	"github.com/riotemergence/godynamicweb/multitenancy"
)

type tenantConnectorHandler struct {
	webApp        *WebApp
	connectorName string
}

func (t tenantConnectorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenantId, result, found := t.webApp.multiTenancySupport.GetTenantIdAndEndpointName(t.connectorName, r)
	if !found {
		http.NotFound(w, r)
		return
	}

	if serverEndpoint, ok := result.(multitenancy.TenantServerEndpoint); ok {
		fmt.Println("serverEndpoint", serverEndpoint)
		handler := t.webApp.serverEndpointsSlots[serverEndpoint.ServerEndpointName].handler
		handler.ServeHTTP(t.webApp, tenantId, w, r)
		return
	}

	if proxyEndpoint, ok := result.(multitenancy.TenantReverseProxyEndpoint); ok {
		fmt.Println("proxyEndpoint", proxyEndpoint)
		http.StripPrefix(proxyEndpoint.StripPrefix, httputil.NewSingleHostReverseProxy(proxyEndpoint.TargetUrl)).ServeHTTP(w, r)
		return
	}

	if fileServerEndpoint, ok := result.(multitenancy.TenantFileServerEndpoint); ok {
		fmt.Println("fileServerEndpoint", fileServerEndpoint)
		http.StripPrefix(fileServerEndpoint.StripPrefix, http.FileServer(http.Dir(fileServerEndpoint.RootFs))).ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
	return
}
