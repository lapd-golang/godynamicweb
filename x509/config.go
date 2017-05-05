package x509

import "fmt"

const TRACE = "github.com/riotemergence/x509"

type X509Config struct {
	PKey []byte `json:"pkey"`
	Cert []byte `json:"cert"`
}

func (x509Config X509Config) Validate() error {
	if x509Config.PKey == nil {
		return fmt.Errorf(TRACE + " X509Config PKey: required")
	}
	if x509Config.Cert == nil {
		return fmt.Errorf(TRACE + " X509Config Cert: required")
	}
	return nil
}
