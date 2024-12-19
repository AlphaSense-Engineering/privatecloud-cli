// Package cloudchecker is the package that contains cloud checking related variables and constants.
package cloudchecker

import (
	"context"
	"errors"
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/mysqlchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/nodegroupchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/oidcchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/smtpchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/ssochecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/storageclasschecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/tlschecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"github.com/charmbracelet/log"
	"go.uber.org/multierr"
	"k8s.io/client-go/kubernetes"
)

var (
	// errFailedToCheckStorageClass is the error that occurs when the storage class is not checked.
	errFailedToCheckStorageClass = errors.New("failed to check storage class")

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
		// logMsgStorageClassChecked is the message that is logged when the storage class is checked.
		logMsgStorageClassChecked = "checked storage class"

		// logMsgNodeGroupsChecked is the message that is logged when the node groups are checked.
		logMsgNodeGroupsChecked = "checked node groups"

		// logMsgNodeGroupsCheckedWithError is the message that is logged when the node groups are checked with an error.
		logMsgNodeGroupsCheckedWithError = "checked node groups; %s"

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

	if _, err := c.storageClassChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(errFailedToCheckStorageClass, err)
	}

	c.logger.Log(log.InfoLevel, logMsgStorageClassChecked)

	if _, err := c.nodeGroupChecker.Handle(ctx); err != nil {
		c.logger.Logf(log.WarnLevel, logMsgNodeGroupsCheckedWithError, err.Error())
	} else {
		c.logger.Log(log.InfoLevel, logMsgNodeGroupsChecked)
	}

	if _, err := c.mySQLChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(errFailedToCheckMySQL, err)
	}

	c.logger.Log(log.InfoLevel, logMsgMySQLChecked)

	if _, err := c.tlsChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(errFailedToCheckTLS, err)
	}

	c.logger.Log(log.InfoLevel, logMsgTLSChecked)

	if _, err := c.smtpChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(errFailedToCheckSMTP, err)
	}

	c.logger.Log(log.InfoLevel, logMsgSMTPChecked)

	if _, err := c.ssoChecker.Handle(ctx); err != nil {
		return nil, multierr.Combine(errFailedToCheckSSO, err)
	}

	c.logger.Log(log.InfoLevel, logMsgSSOChecked)

	jwksURI, err := util.UnwrapValErr[*string](c.oidcChecker.Handle(ctx))
	if err != nil {
		return nil, multierr.Combine(errFailedToCheckOIDCURL, err)
	}

	if jwksURI == nil {
		return nil, nil
	}

	c.logger.Log(log.InfoLevel, logMsgOIDCURLChecked)

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
