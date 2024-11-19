// Package jwtretriever contains the JWT retrieving related variables and constants.
package jwtretriever

import (
	"errors"
)

// ErrNoJWTsRetrieved is an error that occurs when the JWT retriever retrieves no JWTs.
var ErrNoJWTsRetrieved = errors.New("no JWTs retrieved")

const (
	// ServiceAccountsNamespace is the namespace where the Crossplane service accounts are located.
	ServiceAccountsNamespace = "crossplane"
	// TokenExpirationSeconds is the expiration seconds of a single JWT.
	TokenExpirationSeconds = int64(3600)
)
