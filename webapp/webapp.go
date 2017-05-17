package webapp

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/riotemergence/godynamicweb/multitenancy"
	"github.com/riotemergence/godynamicweb/server"
)

const TRACE = "github.com/riotemergence/godynamicweb"

type WebAppStatus int

const (
	StatusUninitialized = iota
	StatusSlotReservation
	StatusRunning
	StatusStopped
)

type MultiTenancyHandler interface {
	ServeHTTP(*WebApp, string, http.ResponseWriter, *http.Request)
}

type MultiTenancyHandlerFunc func(*WebApp, string, http.ResponseWriter, *http.Request)

func (f MultiTenancyHandlerFunc) ServeHTTP(webApp *WebApp, tenantID string, w http.ResponseWriter, r *http.Request) {
	f(webApp, tenantID, w, r)
}

type serverEndpointSlot struct {
	method  string
	handler MultiTenancyHandler
	//Swagger Schema
}

type serverEndpointsSlots map[string]serverEndpointSlot

type parametersNames []string

type clientEndpointSlot struct {
	inputParameters  parametersNames
	outputParameters parametersNames
	//Swagger Schema
}

type clientEndpointsSlots map[string]clientEndpointSlot

type WebApp struct {
	status                          WebAppStatus
	serverEndpointsSlots            serverEndpointsSlots
	httpMethodByServerEndpointSlots map[string]string
	clientEndpointsSlots            clientEndpointsSlots
	server                          *server.Server
	x509Certificates                []tls.Certificate
	x509CertificateBySubjectName    map[string]tls.Certificate
	multiTenancySupport             *multitenancy.MultiTenancySupport
}

func NewWebApp() *WebApp {
	return &WebApp{
		status:                          StatusUninitialized,
		serverEndpointsSlots:            make(serverEndpointsSlots),
		httpMethodByServerEndpointSlots: make(map[string]string),
		clientEndpointsSlots:            make(clientEndpointsSlots),
		server:                          server.NewServer(),
		x509Certificates:                make([]tls.Certificate, 0),
		x509CertificateBySubjectName:    make(map[string]tls.Certificate),
		multiTenancySupport:             multitenancy.NewMultiTenancySupport(),
	}
}

//TODO Swagger
func (webApp *WebApp) SetServerConfigurationSlot(swagger string) error {
	if webApp.status != StatusUninitialized && webApp.status != StatusSlotReservation {
		return fmt.Errorf(TRACE + " WebApp SetConfigurationSlot: statusMustBeStatusUninitializedOrStatusSlotReservation")
	}
	webApp.status = StatusSlotReservation
	return nil
}

//FIXME Swagger
func (webApp *WebApp) AddTenantServerEndpointSlot(serverEndpointName string, httpMethod string, handler MultiTenancyHandlerFunc) error {
	if webApp.status != StatusUninitialized && webApp.status != StatusSlotReservation {
		return fmt.Errorf(TRACE + " WebApp AddServerEndpoint: statusMustBeStatusUninitializedOrStatusSlotReservation")
	}
	webApp.status = StatusSlotReservation

	if serverEndpointName == "" {
		return fmt.Errorf(TRACE + " WebApp AddServerEndpoint serverEndpointName: mustNotBeEmpty")
	}

	_, alreadyExists := webApp.serverEndpointsSlots[serverEndpointName]
	if alreadyExists {
		return fmt.Errorf(TRACE+" WebApp AddServerEndpoint serverEndpointName: alreadyExists \"%s\"", serverEndpointName)
	}

	//TODO interceptors: monitoring, debugging, alarms
	webApp.serverEndpointsSlots[serverEndpointName] = serverEndpointSlot{
		method:  httpMethod,
		handler: handler,
	}
	webApp.httpMethodByServerEndpointSlots[serverEndpointName] = httpMethod

	return nil
}

//FIXME Swagger
func (webApp *WebApp) AddTenantClientEndpointSlot(clientEndpointName string, inputParameters []string, outputParameters []string) error {
	if webApp.status != StatusUninitialized && webApp.status != StatusSlotReservation {
		return fmt.Errorf(TRACE + " WebApp AddClientEndpointSlot: statusMustBeStatusUninitializedOrStatusSlotReservation")
	}
	webApp.status = StatusSlotReservation

	if clientEndpointName == "" {
		return fmt.Errorf(TRACE + " WebApp AddClientEndpoint clientEndpointName: mustNotBeEmpty")
	}

	_, alreadyExists := webApp.clientEndpointsSlots[clientEndpointName]
	if alreadyExists {
		return fmt.Errorf(TRACE + " WebApp AddClientEndpoint clientEndpointName: alreadyExists")
	}
	webApp.clientEndpointsSlots[clientEndpointName] = clientEndpointSlot{
		inputParameters:  inputParameters,
		outputParameters: outputParameters,
	}
	return nil
}

//TODO Swagger
func (webApp *WebApp) SetTenantConfigurationSlot(swagger string) error {
	if webApp.status != StatusUninitialized && webApp.status != StatusSlotReservation {
		return fmt.Errorf(TRACE + " WebApp SetConfigurationSlot: statusMustBeStatusUninitializedOrStatusSlotReservation")
	}
	webApp.status = StatusSlotReservation
	return nil
}

func (webApp *WebApp) CreateServerConnector(connectorName string, connectorConfig server.ConnectorConfig) error {
	if webApp.status != StatusSlotReservation && webApp.status != StatusRunning {
		return fmt.Errorf(TRACE + " WebApp AddServerConnector: statusMustBeStatusSlotReservationOrStatusRunning")
	}
	webApp.status = StatusRunning

	connectorHandler := tenantConnectorHandler{
		webApp:        webApp,
		connectorName: connectorName,
	}
	return webApp.server.AddConnector(
		connectorName,
		connectorConfig,
		connectorHandler,
		getCertificate,
	)
}

func (webApp *WebApp) DeleteServerConnector(connectorName string) error {
	if webApp.status != StatusSlotReservation && webApp.status != StatusRunning {
		return fmt.Errorf(TRACE + " WebApp RemoveServerConnector: statusMustBeStatusSlotReservationOrStatusRunning")
	}
	webApp.status = StatusRunning

	return webApp.server.RemoveConnector(connectorName)
}

func (webApp *WebApp) CreateServerManagementConnector(connectorName string, connectorConfig server.ConnectorConfig) error {
	if webApp.status != StatusSlotReservation && webApp.status != StatusRunning {
		return fmt.Errorf(TRACE + " WebApp AddServerConnector: statusMustBeStatusSlotReservationOrStatusRunning")
	}
	webApp.status = StatusRunning

	mux := mux.NewRouter()
	if err := webApp.server.AddConnector(connectorName, connectorConfig, mux, getCertificate); err != nil {
		return err
	}
	mux.HandleFunc("/", webApp.retrieveServerHandler).Methods(http.MethodGet)
	mux.HandleFunc("/", webApp.deleteServerHandler).Methods(http.MethodDelete)
	mux.HandleFunc("/connectors", webApp.listServerConnectorsHandler).Methods(http.MethodGet)
	mux.HandleFunc("/connectors/{connectorName}", webApp.createOrReplaceServerConnectorHandler).Methods(http.MethodPut)
	mux.HandleFunc("/connectors/{connectorName}", webApp.retrieveServerConnectorHandler).Methods(http.MethodGet)
	mux.HandleFunc("/connectors/{connectorName}", webApp.deleteServerConnectorHandler).Methods(http.MethodDelete)
	mux.HandleFunc("/tenants", webApp.listTenantsHandler).Methods(http.MethodGet)
	mux.HandleFunc("/tenants/{tenantId}", webApp.createOrReplaceTenantHandler).Methods(http.MethodPut)
	mux.HandleFunc("/tenants/{tenantId}", webApp.retrieveTenantHandler).Methods(http.MethodGet)
	mux.HandleFunc("/tenants/{tenantId}", webApp.deleteTenantHandler).Methods(http.MethodDelete)
	mux.HandleFunc("/x509/{x509Cn}", webApp.createOrReplaceTenantHandler).Methods(http.MethodPut)
	mux.HandleFunc("/x509/{x509Cn}", webApp.retrieveTenantHandler).Methods(http.MethodGet)
	mux.HandleFunc("/x509/{x509Cn}", webApp.deleteTenantHandler).Methods(http.MethodDelete)

	return nil
}

func (webApp *WebApp) WaitForTheEnd() error {
	if webApp.status != StatusSlotReservation && webApp.status != StatusRunning {
		return fmt.Errorf(TRACE + " WebApp WaitForTheEnd: statusMustBeStatusSlotReservationOrStatusRunning")
	}
	webApp.status = StatusRunning

	result := webApp.server.WaitForTheEnd()
	webApp.status = StatusStopped
	return result
}

func (webApp *WebApp) AddX509Certificate(privateKeyBytes, certificateChainBytes []byte) error {
	if webApp.status != StatusSlotReservation && webApp.status != StatusRunning {
		return fmt.Errorf(TRACE + " WebApp AddX509Certificate: statusMustBeStatusSlotReservationOrStatusRunning")
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

	commonName := x509Certificate.Subject.CommonName

	if len(commonName) == 0 {
		return fmt.Errorf(TRACE + " WebApp AddX509Certificate: certificateChainBytes CertificateSubjectCommonNameMustNotBeEmpty")
	}

	if _, ok := webApp.x509CertificateBySubjectName[commonName]; ok {
		return fmt.Errorf(TRACE + " WebApp AddX509Certificate: CertificateSubjectCommonNameMustBeUnique")
	}

	for _, subjectAlternativeName := range x509Certificate.DNSNames {
		if len(subjectAlternativeName) == 0 {
			return fmt.Errorf(TRACE + " WebApp AddX509Certificate: certificateChainBytes CertificateSubjectAlternativeNameMustNotBeEmpty")
		}
		if _, ok := webApp.x509CertificateBySubjectName[subjectAlternativeName]; ok {
			return fmt.Errorf(TRACE + " WebApp AddX509Certificate: CertificateSubjectAlternativeNameMustBeUnique")
		}
	}

	//Copy all subject names.
	newX509CertificateBySubjectName := make(map[string]tls.Certificate)
	for k, v := range webApp.x509CertificateBySubjectName {
		newX509CertificateBySubjectName[k] = v
	}

	if len(x509Certificate.Subject.CommonName) > 0 {
		newX509CertificateBySubjectName[x509Certificate.Subject.CommonName] = certificateChainAndPrivateKey
	}
	for _, subjectAlternativeName := range x509Certificate.DNSNames {
		newX509CertificateBySubjectName[subjectAlternativeName] = certificateChainAndPrivateKey
	}

	//webApp.multiTenancySupport.
	return nil
}

func (webApp *WebApp) CreateTenant(tenantID string, config multitenancy.TenantConfig) error {
	if webApp.status != StatusSlotReservation && webApp.status != StatusRunning {
		return fmt.Errorf(TRACE + " WebApp AddTenant: statusMustBeStatusSlotReservationOrStatusRunning")
	}
	webApp.status = StatusRunning

	if err := config.Validate(); err != nil {
		return fmt.Errorf(TRACE+" WebApp CreateTenant config: %s", err)
	}

	for serverEndpointName, serverEndpoint := range *config.ServerEndpoints {
		if _, found := webApp.serverEndpointsSlots[serverEndpointName]; !found {
			return fmt.Errorf(TRACE+" WebApp AddTenant config ServerEndpoints \"%s\": mustMatchRegisteredServerEndpointSlot", serverEndpointName)
		}
		if _, found := webApp.server.RunningEndpointsConnectors[*serverEndpoint.Connector]; !found {
			return fmt.Errorf(TRACE+" WebApp AddTenant config ServerEndpoints \"%s\" Connector: mustExist", *serverEndpoint.Connector)
		}
	}

	for serverEndpointName := range webApp.serverEndpointsSlots {
		if _, found := (*config.ServerEndpoints)[serverEndpointName]; !found {
			return fmt.Errorf(TRACE+" WebApp AddTenant config ServerEndpoints: mustConfigureServerEndpoint \"%s\"", serverEndpointName)
		}
	}

	return webApp.multiTenancySupport.AddTenant(tenantID, config, webApp.httpMethodByServerEndpointSlots)
}

func (webApp *WebApp) DeleteTenant(tenantID string) error {
	if webApp.status != StatusSlotReservation && webApp.status != StatusRunning {
		return fmt.Errorf(TRACE + " WebApp RemoveTenant: statusMustBeStatusSlotReservationOrStatusRunning")
	}
	webApp.status = StatusRunning

	return webApp.multiTenancySupport.RemoveTenant(tenantID)
}

func getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	//clientHello.ServerName
	fmt.Println("TLS: ", clientHello.ServerName)
	return nil, nil
}
