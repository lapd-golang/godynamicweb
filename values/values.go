package values

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// ExistingFile a string representing a existing file path
type ExistingFile string

// String getter for ExistingFile
func (ef *ExistingFile) String() string {
	return string(*ef)
}

// Set setter for ExistingFile
func (ef *ExistingFile) Set(value string) error {
	fi, err := os.Stat(value)
	if os.IsNotExist(err) {
		return fmt.Errorf("\"%s\" file not found", value)
	}
	if fi.IsDir() {
		return fmt.Errorf("\"%s\" must be regular file", value)
	}
	*ef = ExistingFile(value)
	return nil
}

// TokenSigningMethod a string enum representing a Token Signing method
type TokenSigningMethod string

const (
	// HMAC String representing HMAC Token signing Method
	HMAC = "hmac"
	// X509 String representing Certificate Token signing Method (The algorithm info will be extracted from the Certificate)
	X509 = "x509"
)

var validTokenSigningMethod = map[string]string{
	HMAC: "",
	X509: "",
}

// String getter for ExistingFile
func (ts *TokenSigningMethod) String() string {
	return string(*ts)
}

// Set setter for ExistingFile
func (ts *TokenSigningMethod) Set(value string) error {
	_, ok := validTokenSigningMethod[value]
	if ok {
		return nil
	}
	return fmt.Errorf("\"%s\" is not a valid Token Signing Method", value)
}

// AbsoluteURL a URL that checks if it is absolute
type AbsoluteURL string

// String getter for AbsoluteURL
func (u *AbsoluteURL) String() string {
	return string(*u)
}

// Set setter for AbsoluteURL
func (u *AbsoluteURL) Set(value string) error {
	uAsURL, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}
	if !uAsURL.IsAbs() {
		return fmt.Errorf("Must be an absolute URL")
	}
	*u = AbsoluteURL(value)
	return nil
}

// ExistingFileOrAbsoluteURL a URL that checks if it is absolute
type ExistingFileOrAbsoluteURL string

// String getter for ExistingFileOrAbsoluteURL
func (l *ExistingFileOrAbsoluteURL) String() string {
	return string(*l)
}

// Set setter for ExistingFileOrAbsoluteURL
func (l *ExistingFileOrAbsoluteURL) Set(value string) error {
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		uAsURL, err := url.Parse(value)
		if err != nil {
			return fmt.Errorf("invalid URL")
		}
		if !uAsURL.IsAbs() {
			return fmt.Errorf("Must be an absolute URL")
		}
		*l = ExistingFileOrAbsoluteURL(value)
		return nil
	}

	if _, err := os.Stat(value); os.IsNotExist(err) {
		return fmt.Errorf("\"%s\" file not found", value)
	}
	*l = ExistingFileOrAbsoluteURL(value)
	return nil
}
