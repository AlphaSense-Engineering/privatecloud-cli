// Package oidcchecker contains the OIDC checker.
package oidcchecker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
)

var (
	// errOIDCWrongFormat is an error that occurs when the OIDC URL has wrong format.
	errOIDCWrongFormat = errors.New("format of OIDC URL is wrong")

	// errOIDCNon200Response is an error that occurs when the OIDC URL returns non 200 response.
	errOIDCNon200Response = errors.New("non 200 response returned from OIDC URL")

	// errOIDCNoJWKSURI is an error that occurs when the OIDC URL has no jwks_uri field in the response.
	errOIDCNoJWKSURI = errors.New("no jwks_uri field in response returned from OIDC URL")
)

var (
	// awsOIDCRegex is the regex for the OIDC URL for AWS.
	awsOIDCRegex = regexp.MustCompile(
		`^oidc\.eks\.(af|il|ap|ca|eu|me|sa|us|cn|us-gov|us-iso|us-isob)-` +
			`(central|north|(north(?:east|west))|south|south(?:east|west)|east|west)-\d{1}\.amazonaws\.com\/id\/\w+$`,
	)

	// azureOIDCRegex is the regex for the OIDC URL for Azure.
	azureOIDCRegex = regexp.MustCompile(`^https:\/\/.+\.oic\.prod-aks\.azure\.com\/[\w+-]+\/[\w+-]+\/$`)
)

// httpGetter is an interface for abstracting the http.Client.Get method.
//
// There is no real use for this interface besides mocking in tests.
type httpGetter interface {
	// Get sends an HTTP GET request to the specified URL and returns the response.
	Get(string) (*http.Response, error)
}

// OIDCChecker is the OIDC checker.
type OIDCChecker struct {
	// vcloud is the cloud provider.
	vcloud cloud.Cloud
	// envConfig is the environment configuration.
	envConfig *envconfig.EnvConfig
	// httpGetter is the HTTP getter.
	httpGetter httpGetter
}

var _ handler.Handler = &OIDCChecker{}

// Handle is the function that handles the OIDC checking.
//
// The arguments are not used.
// It returns the JWKS URI on success, or an error on failure.
func (c *OIDCChecker) Handle(_ context.Context, _ ...any) ([]any, error) {
	const (
		// httpsScheme is the scheme for the HTTPS URL.
		httpsScheme = "https://"

		// wellKnownEndpoint is the endpoint for the well-known configuration.
		wellKnownEndpoint = "/.well-known/openid-configuration"
	)

	oidcURL := c.envConfig.OIDCURL()

	bytesOIDCURL := []byte(oidcURL)

	if (c.vcloud == cloud.AWS && !awsOIDCRegex.Match(bytesOIDCURL)) ||
		(c.vcloud == cloud.Azure && !azureOIDCRegex.Match(bytesOIDCURL)) {
		return nil, errOIDCWrongFormat
	}

	formattedURL := strings.TrimSuffix(oidcURL, string(constant.HTTPPathSeparator)) + wellKnownEndpoint

	if !strings.HasPrefix(formattedURL, httpsScheme) {
		formattedURL = httpsScheme + formattedURL
	}

	resp, err := c.httpGetter.Get(formattedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, errOIDCNon200Response
	}

	var data struct {
		// JWKSURI is the JWKS URI that is used for validating the JWT.
		JWKSURI *string `json:"jwks_uri,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.JWKSURI == nil {
		return nil, errOIDCNoJWKSURI
	}

	return []any{data.JWKSURI}, nil
}

// New is the function that creates a new OIDC checker.
func New(vcloud cloud.Cloud, envConfig *envconfig.EnvConfig, httpGetter httpGetter) *OIDCChecker {
	return &OIDCChecker{
		vcloud:     vcloud,
		envConfig:  envConfig,
		httpGetter: httpGetter,
	}
}
