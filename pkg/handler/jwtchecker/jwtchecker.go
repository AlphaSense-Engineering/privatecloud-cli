// Package jwtchecker is the package that contains the check functions for JWT.
package jwtchecker

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

var (
	// errJWKSURIRequired is an error that occurs when the JWKS URI is required.
	errJWKSURIRequired = errors.New("JWKS URI is required")

	// errJWTNotValid is an error that occurs when the JWT is not valid.
	errJWTNotValid = errors.New("JWT is not valid")
)

// JWTChecker is the type that contains the check functions for JWT.
type JWTChecker struct {
	// httpClient is the HTTP client.
	httpClient *http.Client
}

var _ handler.Handler = &JWTChecker{}

// Handle is the function that handles the JWT checking.
//
// The first argument is expected to be a pointer to a string representing the JWKS URI.
// The second argument is expected to be a slice of JWTs to be checked.
// It returns nothing on success, or an error on failure.
func (c *JWTChecker) Handle(_ context.Context, args ...any) ([]any, error) {
	jwksURI := handler.ArgAsType[*string](args, 0)

	if jwksURI == nil {
		return nil, errJWKSURIRequired
	}

	jwts := handler.ArgAsType[[]*string](args, 1)

	for _, vjwt := range jwts {
		resp, err := c.httpClient.Get(*jwksURI)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close() // nolint:errcheck

		respJSON, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		jwksKeyfunc, err := keyfunc.NewJWKSetJSON(respJSON)
		if err != nil {
			return nil, err
		}

		parsedJWT, err := jwt.Parse(*vjwt, jwksKeyfunc.Keyfunc)
		if err != nil {
			return nil, err
		}

		if !parsedJWT.Valid {
			return nil, errJWTNotValid
		}
	}

	return nil, nil
}

// New is the function that creates a new JWT checker.
func New(httpClient *http.Client) *JWTChecker {
	return &JWTChecker{httpClient: httpClient}
}
