// Package cloudchecker is the package that contains cloud checking related variables and constants.
package cloudchecker

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/mysqlchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/oidcchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/smtpchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/ssochecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/tlschecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"go.uber.org/multierr"
	"k8s.io/client-go/kubernetes"
)

const (
	// LogMsgJWTsRetrieved is the message that is logged when the JWTs are retrieved.
	LogMsgJWTsRetrieved = "retrieved JWTs"

	// LogMsgJWTsChecked is the message that is logged when the JWTs are checked.
	LogMsgJWTsChecked = "checked JWTs"

	// LogMsgCrossplaneRoleChecked is the message that is logged when the Crossplane role is checked.
	LogMsgCrossplaneRoleChecked = "checked Crossplane role"
)

var (
	// ErrFailedToRetrieveJWTs is the error that occurs when the JWTs are not retrieved.
	ErrFailedToRetrieveJWTs = errors.New("failed to retrieve JWTs")

	// ErrFailedToCheckJWTs is the error that occurs when the JWTs are not checked.
	ErrFailedToCheckJWTs = errors.New("failed to check JWTs")

	// ErrFailedToCheckCrossplaneRole is the error that occurs when the Crossplane role is not checked.
	ErrFailedToCheckCrossplaneRole = errors.New("failed to check Crossplane role")
)

var (
	// errFailedToCheckMySQL is the error that occurs when the MySQL is not checked.
	errFailedToCheckMySQL = errors.New("failed to check MySQL")

	// errFailedToCheckTLS is the error that occurs when the TLS is not checked.
	errFailedToCheckTLS = errors.New("failed to check TLS")

	// errFailedToCheckSMTP is the error that occurs when the SMTP is not checked.
	errFailedToCheckSMTP = errors.New("failed to check SMTP")

	// errFailedToCheckSSO is the error that occurs when the SSO is not checked.
	errFailedToCheckSSO = errors.New("failed to check SSO")

	// errFailedToCheckOIDCURL is the error that occurs when the OIDC URL is not checked.
	errFailedToCheckOIDCURL = errors.New("failed to check OIDC URL")
)

// CloudChecker is the type that contains the infrastructure check functions for cloud.
type CloudChecker struct {
	// vcloud is the cloud provider.
	vcloud cloud.Cloud
	// envConfig is the environment configuration.
	envConfig *envconfig.EnvConfig
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
	// httpClient is the HTTP client.
	httpClient *http.Client

	// mySQLChecker is the MySQL checker.
	mySQLChecker *mysqlchecker.MySQLChecker
	// tlsChecker is the TLS checker.
	tlsChecker *tlschecker.TLSChecker
	// smtpChecker is the SMTP checker.
	smtpChecker *smtpchecker.SMTPChecker
	// ssoChecker is the SSO checker.
	ssoChecker *ssochecker.SSOChecker
	// oidcChecker is the OIDC checker.
	oidcChecker *oidcchecker.OIDCChecker
}

var _ handler.Handler = &CloudChecker{}

// setup is the function that sets up the cloud checker.
func (c *CloudChecker) setup() {
	c.mySQLChecker = mysqlchecker.New(c.clientset)

	c.tlsChecker = tlschecker.New(c.clientset)

	c.smtpChecker = smtpchecker.New(c.clientset)

	c.ssoChecker = ssochecker.New(c.clientset)

	c.oidcChecker = oidcchecker.New(c.vcloud, c.envConfig, c.httpClient)
}

// Handle is the function that handles the infrastructure check.
//
// The arguments are not used.
// It returns the JWKS URI on success, or an error on failure.
//
// nolint:funlen
func (c *CloudChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// logMsgMySQLChecked is the message that is logged when the MySQL is checked.
		logMsgMySQLChecked = "checked MySQL"

		// logMsgTLSChecked is the message that is logged when the TLS is checked.
		logMsgTLSChecked = "checked TLS"

		// logMsgSMTPChecked is the message that is logged when the SMTP is checked.
		logMsgSMTPChecked = "checked SMTP"

		// logMsgSSOChecked is the message that is logged when the SSO is checked.
		logMsgSSOChecked = "checked SSO"

		// logMsgOIDCURLChecked is the message that is logged when the OIDC URL is checked.
		logMsgOIDCURLChecked = "checked OIDC URL"
	)

	if _, err := c.mySQLChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(errFailedToCheckMySQL, err)
	}

	log.Println(logMsgMySQLChecked)

	if _, err := c.tlsChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(errFailedToCheckTLS, err)
	}

	log.Println(logMsgTLSChecked)

	if _, err := c.smtpChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(errFailedToCheckSMTP, err)
	}

	log.Println(logMsgSMTPChecked)

	if _, err := c.ssoChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(errFailedToCheckSSO, err)
	}

	log.Println(logMsgSSOChecked)

	jwksURI, err := util.UnwrapValErr[*string](c.oidcChecker.Handle(ctx))
	if err != nil {
		return nil, multierr.Combine(errFailedToCheckOIDCURL, err)
	}

	if jwksURI == nil {
		return nil, nil
	}

	log.Println(logMsgOIDCURLChecked)

	return []any{jwksURI}, nil
}

// New is the function that creates a new cloud checker.
func New(vcloud cloud.Cloud, envConfig *envconfig.EnvConfig, clientset kubernetes.Interface, httpClient *http.Client) *CloudChecker {
	c := &CloudChecker{
		vcloud:     vcloud,
		envConfig:  envConfig,
		clientset:  clientset,
		httpClient: httpClient,
	}

	c.setup()

	return c
}
