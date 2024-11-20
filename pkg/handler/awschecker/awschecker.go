// Package awschecker is the package that contains the check functions for AWS.
package awschecker

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud/awscloudutil"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/awscrossplanerolechecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/awsjwtretriever"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/jwtchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/mysqlchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/oidcchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/smtpchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/ssochecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/tlschecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.uber.org/multierr"
	"k8s.io/client-go/kubernetes"
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

	// errFailedToRetrieveJWTs is the error that occurs when the JWTs are not retrieved.
	errFailedToRetrieveJWTs = errors.New("failed to retrieve JWTs")

	// errFailedToCheckJWTs is the error that occurs when the JWTs are not checked.
	errFailedToCheckJWTs = errors.New("failed to check JWTs")

	// errFailedToCheckCrossplaneRole is the error that occurs when the Crossplane role is not checked.
	errFailedToCheckCrossplaneRole = errors.New("failed to check Crossplane role")
)

// AWSChecker is the type that contains the infrastructure check functions for AWS.
type AWSChecker struct {
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
	// jwtRetriever is the JWT retriever.
	jwtRetriever *awsjwtretriever.AWSJWTRetriever
	// jwtChecker is the JWT checker.
	jwtChecker *jwtchecker.JWTChecker
}

var _ handler.Handler = &AWSChecker{}

// setup is the function that sets up the AWS checker.
func (c *AWSChecker) setup() {
	c.mySQLChecker = mysqlchecker.New(c.clientset)

	c.tlsChecker = tlschecker.New(c.clientset)

	c.smtpChecker = smtpchecker.New(c.clientset)

	c.ssoChecker = ssochecker.New(c.clientset)

	c.oidcChecker = oidcchecker.New(c.vcloud, c.envConfig, c.httpClient)

	c.jwtRetriever = awsjwtretriever.New(c.clientset)

	c.jwtChecker = jwtchecker.New(c.httpClient)
}

// Handle is the function that handles the infrastructure check.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
//
// nolint:funlen
func (c *AWSChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
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

		// logMsgJWTsRetrieved is the message that is logged when the JWTs are retrieved.
		logMsgJWTsRetrieved = "retrieved JWTs"

		// logMsgJWTsChecked is the message that is logged when the JWTs are checked.
		logMsgJWTsChecked = "checked JWTs"

		// logMsgCrossplaneRoleChecked is the message that is logged when the Crossplane role is checked.
		logMsgCrossplaneRoleChecked = "checked Crossplane role"
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

	log.Println(logMsgOIDCURLChecked)

	jwts, err := util.UnwrapConvertedSliceValErr[any, *string](c.jwtRetriever.Handle(ctx))
	if err != nil {
		return nil, multierr.Combine(errFailedToRetrieveJWTs, err)
	}

	log.Println(logMsgJWTsRetrieved)

	if _, err := c.jwtChecker.Handle(ctx, jwksURI, jwts); err != nil {
		return nil, multierr.Combine(errFailedToCheckJWTs, err)
	}

	log.Println(logMsgJWTsChecked)

	region := c.envConfig.Spec.CloudSpec.CloudZone

	for _, jwt := range jwts {
		stsClient := sts.NewFromConfig(aws.Config{
			Region: region,
		})

		var assumedRole *sts.AssumeRoleWithWebIdentityOutput

		assumedRole, err = stsClient.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
			RoleArn: aws.String(awscloudutil.ARN(
				c.envConfig.Spec.CloudSpec.AWS.AccountID,
				c.envConfig.Spec.ClusterName,
				awscloudutil.ARNTypeRole,
				awscloudutil.CrossplaneRoleName(c.envConfig.Spec.ClusterName),
				nil,
			)),
			RoleSessionName:  aws.String(constant.AppName),
			WebIdentityToken: jwt,
		})
		if err != nil {
			break
		}

		crossplaneRoleChecker := awscrossplanerolechecker.New(c.envConfig, iam.NewFromConfig(aws.Config{
			Region: region,
			Credentials: credentials.NewStaticCredentialsProvider(
				*assumedRole.Credentials.AccessKeyId,
				*assumedRole.Credentials.SecretAccessKey,
				*assumedRole.Credentials.SessionToken,
			),
		}))

		if _, err := crossplaneRoleChecker.Handle(ctx); err != nil {
			break
		}
	}

	if err != nil {
		return nil, multierr.Combine(errFailedToCheckCrossplaneRole, err)
	}

	log.Println(logMsgCrossplaneRoleChecked)

	return []any{}, nil
}

// New is the function that creates a new AWS checker.
func New(vcloud cloud.Cloud, envConfig *envconfig.EnvConfig, clientset kubernetes.Interface, httpClient *http.Client) *AWSChecker {
	c := &AWSChecker{
		vcloud:     vcloud,
		envConfig:  envConfig,
		clientset:  clientset,
		httpClient: httpClient,
	}

	c.setup()

	return c
}
