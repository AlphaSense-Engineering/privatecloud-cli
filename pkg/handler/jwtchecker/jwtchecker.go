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

// LogMsgJWTsChecked is the message that is logged when the JWTs are checked.
const LogMsgJWTsChecked = "checked JWTs"

var (
	// ErrFailedToCheckJWTs is the error that occurs when the JWTs are not checked.
	ErrFailedToCheckJWTs = errors.New("failed to check JWTs")

	// errJWTNotValid is an error that occurs when the JWT is not valid.
	errJWTNotValid = errors.New("jwt is not valid")
)

// JWTChecker is the type that contains the check functions for JWT.
type JWTChecker struct {
	// httpClient is the HTTP client.
	httpClient *http.Client
	// jwksURI is the JWKS URI.
	jwksURI *string
}

var _ handler.Handler = &JWTChecker{}

// Handle is the function that handles the JWT checking.
//
// The argument is expected to be a slice of JWTs to be checked.
// It returns nothing on success, or an error on failure.
func (c *JWTChecker) Handle(_ context.Context, args ...any) ([]any, error) {
	jwts := handler.ArgAsType[[]*string](args, 0)

	for _, vjwt := range jwts {
		resp, err := c.httpClient.Get(*c.jwksURI)
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
func New(httpClient *http.Client, jwksURI *string) *JWTChecker {
	return &JWTChecker{httpClient: httpClient, jwksURI: jwksURI}
}
