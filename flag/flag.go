//TODO Remove common microservices flags to a component

package flag

import (
	"flag"
	"fmt"
	"os"

	"github.com/riotemergence/godynamicweb/values"
)

const (
	registrationEndpointFlag = "registry-base-url"
	specsLocationFlag        = "specs-location"
)

// Flag all the command line options are aggregated in this type.
var Flag Flags

// Flags all the command line options are aggregated in this type.
type Flags struct {
	registrationEndpoint values.AbsoluteURL
	specsLocation        values.ExistingFileOrAbsoluteURL
}

func DefineFlags() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... \n", os.Args[0])
		fmt.Fprint(os.Stderr, "Starts a standalone HTTP/S server that awswer to OAuth2 authorization and token requests\n\n")
		flag.PrintDefaults()
	}
	flag.Var(&Flag.registrationEndpoint, registrationEndpointFlag, "The Service Catalog Registration Endpoint (Absolute HTTP URL)")
	flag.Var(&Flag.specsLocation, specsLocationFlag, "The Location where the specs can be found. (Existing file or Absolute HTTP URL)")
	// flag.Var(&Flag.errorURL, errorURLFlag, "The Error Redirect URL")
	// flag.Var(&Flag.errorURL, errorURLFlag, "The Error Redirect URL")
	// flag.Var(&Flag.loginURL, loginURLFlag, "The Login Redirect URL")
	// flag.Var(&Flag.port, portFlag, "The server port. (1-65535) (default 80 for HTTP, 443 for HTTPS)")
	// flag.Var(&Flag.serverX509CertificateFilePath, serverX509CertificateFlag, "The HTTPS server certificate file path. (existing file)")
	// flag.Var(&Flag.serverX509PrivateKeyFilePath, serverX509PrivateKeyFlag, "The HTTPS server private key file path. (existing file)")
	// flag.Var(&Flag.tokenSigningMethod, tokenSigningMethodFlag, "The Token Signing method. (\"hmac\", \"x509\") (default \"x509\")")
	// flag.Var(&Flag.tokenX509CertificateFilePath, tokenX509CertificateFlag, "The Token Signing certificate file path. (existing file)")
	// flag.Var(&Flag.tokenX509PrivateKeyFilePath, tokenX509PrivateKeyFlag, "The Token signing private key file path. (existing file)")
	// flag.BoolVar(&Flag.useTLS, useTLSFlag, true, "Define if server uses HTTP or HTTPS. HTTPS is mandatory for OpenID/OAuth2 Compliance. (boolean)")
}

func CheckFlags() error {
	if Flag.specsLocation.String() == "" {
		return fmt.Errorf("invalid value for flag -%s: an existing file or absolute URL is required", specsLocationFlag)
	}
	return nil
}

func ParseFlags() {
	flag.Parse()
}
