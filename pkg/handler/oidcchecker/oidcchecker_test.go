// Package oidcchecker contains the OIDC checker.
package oidcchecker

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"github.com/stretchr/testify/assert"
)

// mockHTTPGetter is a mock implementation of the httpGetter interface.
type mockHTTPGetter struct {
	// statusCode is the status code to return.
	statusCode int
	// bodyString is the body to return.
	bodyString string
}

var _ httpGetter = &mockHTTPGetter{}

// Get is a mock implementation of the Get method.
func (m *mockHTTPGetter) Get(_ string) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.bodyString)),
	}, nil
}

// TestOIDCChecker_Handle tests the OIDCChecker.Handle method.
//
// nolint:funlen
func TestOIDCChecker_Handle(t *testing.T) {
	const (
		// validAWSURL is a valid AWS OIDC URL.
		validAWSURL = "oidc.eks.us-west-2.amazonaws.com/id/foo"

		// validAzureURL is a valid Azure OIDC URL.
		validAzureURL = "https://example.oic.prod-aks.azure.com/foo/bar/"

		// invalidOIDCURL is an invalid OIDC URL.
		invalidOIDCURL = "invalid"

		// validBodyString is a valid body string.
		validBodyString = `{"jwks_uri": "irrelevant"}`

		// emptyJSONBodyString is an empty JSON body string.
		emptyJSONBodyString = `{}`

		// emptyStringBodyString is an empty string body string.
		emptyStringBodyString = ""

		// irrelevant is an irrelevant value.
		irrelevant = "irrelevant"
	)

	testCases := []struct {
		name        string
		oidcURL     string
		cloud       cloud.Cloud
		statusCode  int
		bodyString  string
		wantJWKSURI []any
		wantErr     error
	}{
		{
			name:        "Valid AWS OIDC URL",
			oidcURL:     validAWSURL,
			cloud:       cloud.AWS,
			statusCode:  http.StatusOK,
			bodyString:  validBodyString,
			wantJWKSURI: []any{util.Ref(irrelevant)},
			wantErr:     nil,
		},
		{
			name:        "Valid Azure OIDC URL",
			oidcURL:     validAzureURL,
			cloud:       cloud.Azure,
			statusCode:  http.StatusOK,
			bodyString:  validBodyString,
			wantJWKSURI: []any{util.Ref(irrelevant)},
			wantErr:     nil,
		},
		{
			name:       "Valid AWS OIDC URL with non-200 status code",
			oidcURL:    validAWSURL,
			cloud:      cloud.AWS,
			statusCode: http.StatusTeapot,
			bodyString: validBodyString,
			wantErr:    errOIDCNon200Response,
		},
		{
			name:       "Valid Azure OIDC URL with non-200 status code",
			oidcURL:    validAzureURL,
			cloud:      cloud.Azure,
			statusCode: http.StatusTeapot,
			bodyString: validBodyString,
			wantErr:    errOIDCNon200Response,
		},
		{
			name:       "Valid AWS OIDC URL without JWKS URI field in the response",
			oidcURL:    validAWSURL,
			cloud:      cloud.AWS,
			statusCode: http.StatusOK,
			bodyString: emptyJSONBodyString,
			wantErr:    errOIDCNoJWKSURI,
		},
		{
			name:       "Valid Azure OIDC URL without JWKS URI field in the response",
			oidcURL:    validAzureURL,
			cloud:      cloud.Azure,
			statusCode: http.StatusOK,
			bodyString: emptyJSONBodyString,
			wantErr:    errOIDCNoJWKSURI,
		},
		{
			name:       "Valid AWS OIDC URL with empty string body",
			oidcURL:    validAWSURL,
			cloud:      cloud.AWS,
			statusCode: http.StatusOK,
			bodyString: emptyStringBodyString,
			wantErr:    io.EOF,
		},
		{
			name:       "Valid Azure OIDC URL with empty string body",
			oidcURL:    validAzureURL,
			cloud:      cloud.Azure,
			statusCode: http.StatusOK,
			bodyString: emptyStringBodyString,
			wantErr:    io.EOF,
		},
		{
			name:       "Invalid AWS OIDC URL",
			oidcURL:    invalidOIDCURL,
			cloud:      cloud.AWS,
			statusCode: http.StatusNotFound,
			bodyString: emptyJSONBodyString,
			wantErr:    errOIDCWrongFormat,
		},
		{
			name:       "Invalid Azure OIDC URL",
			oidcURL:    invalidOIDCURL,
			cloud:      cloud.Azure,
			statusCode: http.StatusNotFound,
			bodyString: emptyJSONBodyString,
			wantErr:    errOIDCWrongFormat,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envCfg := &envconfig.EnvConfig{
				Spec: envconfig.Spec{
					CloudSpec: envconfig.CloudSpec{
						Provider: string(tc.cloud),
					},
				},
			}

			if tc.cloud == cloud.AWS {
				envCfg.Spec.CloudSpec.AWS = &envconfig.AWSSpec{
					OIDCURL: tc.oidcURL,
				}
			} else if tc.cloud == cloud.Azure {
				envCfg.Spec.CloudSpec.Azure = &envconfig.AzureSpec{
					OIDCURL: tc.oidcURL,
				}
			}

			oidcChecker := New(
				tc.cloud,
				envCfg,
				&mockHTTPGetter{
					statusCode: tc.statusCode,
					bodyString: tc.bodyString,
				},
			)

			gotJWKSURI, gotErr := oidcChecker.Handle(context.TODO())

			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, gotErr, "got %v, want %v", gotErr, tc.wantErr)
			}

			if tc.wantJWKSURI != nil {
				assert.Equal(t, tc.wantJWKSURI, gotJWKSURI, "got %v, want %v", gotJWKSURI, tc.wantJWKSURI)
			}
		})
	}
}
