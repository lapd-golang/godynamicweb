package x509

import "crypto/tls"

type X509Config struct {
	PKey []byte `json:"pkey"`
	Cert []byte `json:"cert"`
}

func (x509Config X509Config) Validate() error {
	_, err := tls.X509KeyPair(x509Config.Cert, x509Config.PKey)
	if err != nil {
		return err
	}
	return nil
}
