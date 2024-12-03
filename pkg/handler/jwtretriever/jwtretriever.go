// Package jwtretriever contains the JWT retrieving related variables and constants.
package jwtretriever

import (
	"errors"
)

// LogMsgJWTsRetrieved is the message that is logged when the JWTs are retrieved.
const LogMsgJWTsRetrieved = "retrieved JWTs"

var (
	// ErrFailedToRetrieveJWTs is the error that occurs when the JWTs are not retrieved.
	ErrFailedToRetrieveJWTs = errors.New("failed to retrieve JWTs")

	// ErrNoJWTsRetrieved is an error that occurs when the JWT retriever retrieves no JWTs.
	ErrNoJWTsRetrieved = errors.New("no JWTs retrieved")
)

const (
	// ServiceAccountsNamespace is the namespace where the Crossplane service accounts are located.
	ServiceAccountsNamespace = "crossplane"

	// TokenExpirationSeconds is the expiration seconds of a single JWT.
	TokenExpirationSeconds = int64(3600)
)
