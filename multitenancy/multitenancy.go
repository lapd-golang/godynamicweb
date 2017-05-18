package multitenancy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"crypto/tls"
	"crypto/x509"

	"github.com/riotemergence/godynamicweb/mux"
)

const TRACE = "github.com/riotemergence/godynamicweb/multitenancy"

type TenantServerEndpoint struct {
	TenantID           string
	ServerEndpointName string
}

type TenantReverseProxyEndpoint struct {
	TenantID    string
	StripPrefix string
	TargetUrl   *url.URL
}

type TenantFileServerEndpoint struct {
	TenantID    string
	StripPrefix string
	RootFs      string
	DirListing  bool
}

type MultiTenancySupport struct {
	Config     MultiTenancyConfig
	MuxCatalog *mux.MuxCatalog
	//	x509Certificates            []tls.Certificate
	x509CertificateBySubjectName map[string]tls.Certificate
}

func NewMultiTenancySupport() *MultiTenancySupport {
	multiTenancy := &MultiTenancySupport{
		Config: MultiTenancyConfig{
			Tenants: make(TenantsConfig),
		},
		MuxCatalog: mux.NewMuxCatalog(),
	}
	return multiTenancy
}

//TODO Check if connector use tls if https url is used
func (m *MultiTenancySupport) AddTenant(tenantID string, config TenantConfig, httpMethodByServerEndpointName map[string]string) error {
	if tenantID == "" {
		return fmt.Errorf(TRACE + " MultiTenancySupport AddTenant tenantID: mustNotBeEmpty")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf(TRACE+" MultiTenancySupport AddTenant config: %s", err)
	}

	if httpMethodByServerEndpointName == nil {
		return fmt.Errorf(TRACE + " MultiTenancySupport AddTenant httpMethodByServerEndpointName: mustNotBeEmpty")
	}

	if _, found := m.Config.Tenants[tenantID]; found {
		return fmt.Errorf(TRACE+" MultiTenancySupport AddTenant tenantID: mustNotExist \"%s\"", tenantID)
	}

	if config.X509 != nil {
		for _, x509 := range config.X509 {
			if err := x509.Validate(); err != nil {
				return err
			}
		}
	}

	certificateChainAndPrivateKey, err := tls.X509KeyPair(certificateChainBytes, privateKeyBytes)
	if err != nil {
		return err
	}

	certificateBytes := certificateChainAndPrivateKey.Certificate[0]
	x509Certificate, err := x509.ParseCertificate(certificateBytes)
	if err != nil {
		return err
	}

	//config.X509.

	if _, ok := m.x509CertificateBySubjectName(commonName); ok {
		return fmt.Errorf(TRACE + " WebApp AddX509Certificate: CertificateSubjectCommonNameMustBeUnique")
	}

	for _, subjectAlternativeName := range x509Certificate.DNSNames {
		if len(subjectAlternativeName) == 0 {
			return fmt.Errorf(TRACE + " WebApp AddX509Certificate: certificateChainBytes CertificateSubjectAlternativeNameMustNotBeEmpty")
		}
		if _, ok := webApp.x509CertificateBySubjectName[subjectAlternativeName]; ok {
			return fmt.Errorf(TRACE + " WebApp AddX509Certificate: CertificateSubjectAlternativeNameMustBeUnique")
		}
		if _, ok := webApp.multiTenancySupport.GetCertificateChainAndPrivateKeyBySubjectName(subjectAlternativeName); ok {
			return fmt.Errorf(TRACE + " WebApp AddX509Certificate: CertificateSubjectCommonNameMustBeUnique")
		}
	}

	TempMuxCatalog := *m.MuxCatalog

	for serverEndpointName, serverEndpointValue := range *config.ServerEndpoints {
		serverEndpointURL, err := url.Parse(string(*serverEndpointValue.Url))
		if err != nil {
			return err
		}

		httpMethod, found := httpMethodByServerEndpointName[serverEndpointName]
		if !found {
			return fmt.Errorf(TRACE + " MultiTenancySupport AddTenant config ServerEndpoints \"%s\" : mustExistsInHttpMethodByServerEndpointName")
		}

		err = TempMuxCatalog.Add(
			*serverEndpointValue.Connector,
			serverEndpointURL.Scheme,
			serverEndpointURL.Host,
			serverEndpointURL.Path,
			httpMethod,
			TenantServerEndpoint{
				tenantID,
				serverEndpointName,
			},
		)
		if err != nil {
			return fmt.Errorf(TRACE+" MultiTenancySupport AddTenant config url : mustNotConflictWithExistingUrl \"%s\"", *serverEndpointValue.Url)
		}
	}

	if config.ReverseProxyEndpoints != nil {
		for _, reverseProxyEndpoint := range *config.ReverseProxyEndpoints {
			reverseProxySourceUrl, err := url.Parse(string(*reverseProxyEndpoint.Url))
			if err != nil {
				return err
			}

			reverseProxyTargetUrl, err := url.Parse(string(*reverseProxyEndpoint.TargetUrl))
			if err != nil {
				return err
			}

			for _, method := range *reverseProxyEndpoint.Methods {
				err := TempMuxCatalog.Add(
					*reverseProxyEndpoint.Connector,
					reverseProxySourceUrl.Scheme,
					reverseProxySourceUrl.Host,
					reverseProxySourceUrl.Path,
					method,
					TenantReverseProxyEndpoint{
						tenantID,
						saneDirTerminator(reverseProxySourceUrl.Path),
						reverseProxyTargetUrl,
					},
				)
				if err != nil {
					return fmt.Errorf(TRACE + " MultiTenancySupport AddTenant config ReverseProxies Url: mustNotConflictWithExistingUrl")
				}
			}
		}
	}

	if config.FileServerEndpoints != nil {
		for _, fileServerEndpoint := range *config.FileServerEndpoints {
			fileServerUrl, err := url.Parse(string(*fileServerEndpoint.Url))
			if err != nil {
				return err
			}

			err = TempMuxCatalog.Add(
				*fileServerEndpoint.Connector,
				fileServerUrl.Scheme,
				fileServerUrl.Host,
				fileServerUrl.Path,
				http.MethodGet,
				TenantFileServerEndpoint{
					tenantID,
					saneDirTerminator(fileServerUrl.Path),
					string(*fileServerEndpoint.RootFs),
					fileServerEndpoint.DirListing != nil && *fileServerEndpoint.DirListing,
				},
			)
			if err != nil {
				return fmt.Errorf(TRACE + " MultiTenancySupport AddTenant config FileServerEndpoints Url: mustNotConflictWithExistingUrl")
			}
		}
	}

	m.MuxCatalog = &TempMuxCatalog
	m.Config.Tenants[tenantID] = config
	return nil
}

func (m *MultiTenancySupport) RemoveTenant(tenantID string) error {
	if _, ok := m.Config.Tenants[tenantID]; !ok {
		return fmt.Errorf(TRACE+" MultiTenancySupport RemoveTenant tenantID: mustExists \"%s\"", tenantID)
	}

	removeWhen := func(muxEntry mux.MuxEntry) bool {
		serverEndpoint, ok := muxEntry.Value.(TenantServerEndpoint)
		if ok && serverEndpoint.TenantID == tenantID {
			return true
		}
		proxyEndpoint, ok := muxEntry.Value.(TenantReverseProxyEndpoint)
		return ok && proxyEndpoint.TenantID == tenantID

	}
	m.MuxCatalog.RemoveAll(removeWhen)

	return nil
}

func (m *MultiTenancySupport) GetTenantIdAndEndpointName(connectorName string, r *http.Request) (tenantID string, result interface{}, found bool) {
	muxEntry, found := m.MuxCatalog.GetWithRequest(connectorName, r)
	if !found {
		return "", nil, false
	}

	fmt.Println(muxEntry.Key, muxEntry.Value)
	serverEndpoint, ok := muxEntry.Value.(TenantServerEndpoint)
	if ok {
		return serverEndpoint.TenantID, serverEndpoint, true
	}

	proxyEndpoint, ok := muxEntry.Value.(TenantReverseProxyEndpoint)
	if ok {
		return proxyEndpoint.TenantID, proxyEndpoint, true
	}

	fsEndpoint, ok := muxEntry.Value.(TenantFileServerEndpoint)
	if ok {
		return fsEndpoint.TenantID, fsEndpoint, true
	}

	return "", nil, false

}

func (m *MultiTenancySupport) GetCertificateChainAndPrivateKeyBySubjectName(subjectName string) (tls.Certificate, bool) {
	certificate, ok := m.x509CertificateBySubjectName[subjectName]
	return certificate, ok
}

func saneDirTerminator(s string) string {
	if strings.HasSuffix(s, "/*") {
		return s[0 : len(s)-1]
	}
	return s
}
