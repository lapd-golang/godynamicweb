package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

const TRACE = "github.com/riotemergence/godynamicweb/server"

type Server struct {
	Config                     *ServerConfig
	DoneAndErrorChannel        chan error
	RunningEndpointsConnectors map[string]*Connector
}

type Connector struct {
	Config              ConnectorConfig
	DoneAndErrorChannel chan error
	Listener            net.Listener
	Mux                 http.Handler
	Server              *http.Server
}

// // Setup HTTPS client
// tlsConfig := &tls.Config{
// 	ClientCAs: caCertPool,
// 	// NoClientCert
// 	// RequestClientCert
// 	// RequireAnyClientCert
// 	// VerifyClientCertIfGiven
// 	// RequireAndVerifyClientCert
// 	ClientAuth: tls.RequireAndVerifyClientCert,
// }
// tlsConfig.BuildNameToCertificate()

// server := &http.Server{
// 	Addr:      ":8080",
// 	TLSConfig: tlsConfig,
// }

// server.ListenAndServeTLS("selfsigned.crt", "selfsigned.key") //private cert

// ief, err := net.InterfaceByName("eth1")
// if err != nil {
// 	log.Fatal(err)
// }
// addrs, err := ief.Addrs()
// if err != nil {
// 	log.Fatal(err)
// }

// tcpAddr := &net.TCPAddr{
// 	IP: addrs[0].(*net.IPNet).IP,
// }

//defer endpointServer.Close()

func NewServer() *Server {
	server := &Server{
		Config: NewServerConfig(),
		RunningEndpointsConnectors: make(map[string]*Connector),
		DoneAndErrorChannel:        make(chan error),
	}

	return server
}

func (s *Server) AddConnector(connectorName string, config ConnectorConfig, mux http.Handler, getCertificateFunc func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error)) error {
	if err := config.Validate(); err != nil {
		return err
	}
	_, ok := (*s.Config.Connectors)[connectorName]
	if ok {
		return fmt.Errorf(TRACE + " AddConnector connectorName: alreadyExists")
	}

	var tlsConfig *tls.Config = nil

	connectorServerAddr := string(*config.BindAddress) + ":" + strconv.Itoa(int(*config.Port))
	connectorServer := &http.Server{
		Addr:           connectorServerAddr,
		Handler:        mux,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		TLSConfig:      tlsConfig,
	}

	var connectorServerListener net.Listener
	{
		var err error
		if config.TLS != nil && *config.TLS {
			tlsConfig = &tls.Config{
				GetCertificate: getCertificateFunc,
			}
			connectorServerListener, err = tls.Listen("tcp", connectorServerAddr, tlsConfig)
			if err != nil {
				return err
			}
		} else {
			connectorServerListener, err = net.Listen("tcp", connectorServerAddr)
			if err != nil {
				return err
			}
		}
	}
	doneAndErrorChannel := make(chan error)

	connector := &Connector{
		Config:              config,
		DoneAndErrorChannel: doneAndErrorChannel,
		Listener:            connectorServerListener,
		Mux:                 mux,
		Server:              connectorServer,
	}

	go func() {
		err := connectorServer.Serve(connectorServerListener)
		if err != nil {
			connector.DoneAndErrorChannel <- err
			return
		}
		connector.DoneAndErrorChannel <- nil
	}()

	s.RunningEndpointsConnectors[connectorName] = connector
	(*s.Config.Connectors)[connectorName] = config
	return nil
}

func (s *Server) RemoveConnector(connectorName string) error {
	c, ok := s.RunningEndpointsConnectors[connectorName]
	if !ok {
		return fmt.Errorf("RemoveConnector connectorName notFound")
	}

	go func() {
		err := c.Listener.Close()
		if err != nil {
			c.DoneAndErrorChannel <- err
		}
	}()
	err := <-c.DoneAndErrorChannel
	errOpError, ok := err.(*net.OpError)
	if ok && errOpError.Op != "accept" {
		return err
	}
	delete(s.RunningEndpointsConnectors, connectorName)
	delete(*s.Config.Connectors, connectorName)

	// if len(s.RunningEndpointsConnectors) == 0 {

	// }
	return nil
}

func (s *Server) Stop() {
	for k := range s.RunningEndpointsConnectors {
		s.RemoveConnector(k)
	}
	s.DoneAndErrorChannel <- nil
}

func (s *Server) WaitForTheEnd() error {
	return <-s.DoneAndErrorChannel
}

func (server *Server) String() string {
	return "Up since the dawn of computers"
}
