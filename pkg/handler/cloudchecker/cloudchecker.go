// Package cloudchecker is the package that contains cloud checking related variables and constants.
package cloudchecker

import (
	"context"
	"errors"
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/mysqlchecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/nodegroupchecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/oidcchecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/postgresqlchecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/smtpchecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/ssochecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/storageclasschecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/tlschecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	"github.com/charmbracelet/log"
	"go.uber.org/multierr"
	"k8s.io/client-go/kubernetes"
)

var (
	// ErrFailedToCheckStorageClass is the error that occurs when the storage class is not checked.
	ErrFailedToCheckStorageClass = errors.New("failed to check storage class")

	// ErrFailedToCheckMySQL is the error that occurs when the MySQL is not checked.
	ErrFailedToCheckMySQL = errors.New("failed to check MySQL")

	// ErrFailedToCheckPostgreSQL is the error that occurs when the PostgreSQL is not checked.
	ErrFailedToCheckPostgreSQL = errors.New("failed to check PostgreSQL")

	// ErrFailedToCheckTLS is the error that occurs when the TLS is not checked.
	ErrFailedToCheckTLS = errors.New("failed to check TLS")

	// ErrFailedToCheckSMTP is the error that occurs when the SMTP is not checked.
	ErrFailedToCheckSMTP = errors.New("failed to check SMTP")

	// ErrFailedToCheckSSO is the error that occurs when the SSO is not checked.
	ErrFailedToCheckSSO = errors.New("failed to check SSO")

	// ErrFailedToCheckOIDCURL is the error that occurs when the OIDC URL is not checked.
	ErrFailedToCheckOIDCURL = errors.New("failed to check OIDC URL")
)

// CloudChecker is the type that contains the infrastructure check functions for cloud.
type CloudChecker struct {
	// logger is the logger.
	logger *log.Logger
	// vcloud is the cloud provider.
	vcloud cloud.Cloud
	// envConfig is the environment configuration.
	envConfig *envconfig.EnvConfig
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
	// httpClient is the HTTP client.
	httpClient *http.Client

	// storageClassChecker is the storage class checker.
	storageClassChecker *storageclasschecker.StorageClassChecker
	// nodeGroupChecker is the node group checker.
	nodeGroupChecker *nodegroupchecker.NodeGroupChecker

	// mySQLChecker is the MySQL checker.
	mySQLChecker *mysqlchecker.MySQLChecker
	// postgresqlChecker is the PostgreSQL checker.
	postgresqlChecker *postgresqlchecker.PostgreSQLChecker
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
	c.storageClassChecker = storageclasschecker.New(c.clientset)

	c.nodeGroupChecker = nodegroupchecker.New(c.clientset)

	c.mySQLChecker = mysqlchecker.New(c.clientset)

	c.postgresqlChecker = postgresqlchecker.New(c.clientset)

	c.tlsChecker = tlschecker.New(c.clientset)

	c.smtpChecker = smtpchecker.New(c.clientset)

	c.ssoChecker = ssochecker.New(c.clientset)

	c.oidcChecker = oidcchecker.New(c.vcloud, c.envConfig, c.httpClient)
}

// Handle is the function that handles the infrastructure check.
//
// Checks in this function are ordered in the same way as they are listed at https://developer.alpha-sense.com/enterprise/technical-requirements.
//
// The arguments are not used.
// It returns the JWKS URI on success, or an error on failure.
//
// nolint:funlen
func (c *CloudChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// logMsgStorageClassCheckedSuccessfully is the message that is logged when the storage class is checked successfully.
		logMsgStorageClassCheckedSuccessfully = "checked storage class successfully"

		// logMsgNodeGroupsCheckedSuccessfully is the message that is logged when the node groups are checked successfully.
		logMsgNodeGroupsCheckedSuccessfully = "checked node groups successfully"

		// logMsgNodeGroupsCheckedWarn is the message that is logged when the node groups are checked with a warning.
		logMsgNodeGroupsCheckedWarn = "checked node groups; %s"

		// logMsgMySQLCheckedSuccessfully is the message that is logged when the MySQL is checked successfully.
		logMsgMySQLCheckedSuccessfully = "checked MySQL successfully"

		// logMsgPostgreSQLCheckedSuccessfully is the message that is logged when the PostgreSQL is checked successfully.
		logMsgPostgreSQLCheckedSuccessfully = "checked PostgreSQL successfully"

		// logMsgTLSCheckedSuccessfully is the message that is logged when the TLS is checked successfully.
		logMsgTLSCheckedSuccessfully = "checked TLS successfully"

		// logMsgSMTPCheckedSuccessfully is the message that is logged when the SMTP is checked successfully.
		logMsgSMTPCheckedSuccessfully = "checked SMTP successfully"

		// logMsgSSOCheckedSuccessfully is the message that is logged when the SSO is checked successfully.
		logMsgSSOCheckedSuccessfully = "checked SSO successfully"

		// logMsgOIDCURLCheckedSuccessfully is the message that is logged when the OIDC URL is checked successfully.
		logMsgOIDCURLCheckedSuccessfully = "checked OIDC URL successfully"
	)

	if _, err := c.storageClassChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(ErrFailedToCheckStorageClass, err)
	}

	c.logger.Info(logMsgStorageClassCheckedSuccessfully)

	if _, err := c.nodeGroupChecker.Handle(ctx); err != nil {
		c.logger.Logf(log.WarnLevel, logMsgNodeGroupsCheckedWarn, err.Error())
	} else {
		c.logger.Info(logMsgNodeGroupsCheckedSuccessfully)
	}

	if _, err := c.mySQLChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(ErrFailedToCheckMySQL, err)
	}

	c.logger.Info(logMsgMySQLCheckedSuccessfully)

	if _, err := c.postgresqlChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(ErrFailedToCheckPostgreSQL, err)
	}

	c.logger.Info(logMsgPostgreSQLCheckedSuccessfully)

	if _, err := c.tlsChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(ErrFailedToCheckTLS, err)
	}

	c.logger.Info(logMsgTLSCheckedSuccessfully)

	if _, err := c.smtpChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(ErrFailedToCheckSMTP, err)
	}

	c.logger.Info(logMsgSMTPCheckedSuccessfully)

	if _, err := c.ssoChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(ErrFailedToCheckSSO, err)
	}

	c.logger.Info(logMsgSSOCheckedSuccessfully)

	jwksURI, err := util.UnwrapValErr[*string](c.oidcChecker.Handle(ctx))
	if err != nil {
		return nil, multierr.Combine(ErrFailedToCheckOIDCURL, err)
	}

	if jwksURI == nil {
		return nil, nil
	}

	c.logger.Info(logMsgOIDCURLCheckedSuccessfully)

	return []any{jwksURI}, nil
}

// New is the function that creates a new CloudChecker.
func New(logger *log.Logger, vcloud cloud.Cloud, envConfig *envconfig.EnvConfig, clientset kubernetes.Interface, httpClient *http.Client) *CloudChecker {
	c := &CloudChecker{
		logger:     logger,
		vcloud:     vcloud,
		envConfig:  envConfig,
		clientset:  clientset,
		httpClient: httpClient,
	}

	c.setup()

	return c
}
