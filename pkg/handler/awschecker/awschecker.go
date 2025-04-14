// Package awschecker is the package that contains the check functions for AWS.
package awschecker

import (
	"context"
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/cloud/awscloudutil"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/awscrossplanerolechecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/awsjwtretriever"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/crossplanerolechecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/jwtchecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/jwtretriever"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/charmbracelet/log"
	"go.uber.org/multierr"
	"k8s.io/client-go/kubernetes"
)

// AWSChecker is the type that contains the infrastructure check functions for AWS.
type AWSChecker struct {
	// logger is the logger.
	logger *log.Logger
	// envConfig is the environment configuration.
	envConfig *envconfig.EnvConfig
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
	// httpClient is the HTTP client.
	httpClient *http.Client
	// jwksURI is the JWKS URI.
	jwksURI *string

	// jwtRetriever is the JWT retriever.
	jwtRetriever *awsjwtretriever.AWSJWTRetriever
	// jwtChecker is the JWT checker.
	jwtChecker *jwtchecker.JWTChecker
}

var _ handler.Handler = &AWSChecker{}

// setup is the function that sets up the AWS checker.
func (c *AWSChecker) setup() {
	c.jwtRetriever = awsjwtretriever.New(c.clientset)

	c.jwtChecker = jwtchecker.New(c.httpClient, c.jwksURI)
}

// Handle is the function that handles the infrastructure check.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
//
// nolint:funlen
func (c *AWSChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	jwts, err := util.ConvertSliceErr[any, *string](c.jwtRetriever.Handle(ctx))
	if err != nil {
		return nil, multierr.Combine(jwtretriever.ErrFailedToRetrieveJWTs, err)
	}

	c.logger.Debug(jwtretriever.LogMsgJWTsRetrieved)

	if _, err := c.jwtChecker.Handle(ctx, jwts); err != nil {
		return nil, multierr.Combine(jwtchecker.ErrFailedToCheckJWTs, err)
	}

	c.logger.Debug(jwtchecker.LogMsgJWTsChecked)

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

		crossplaneRoleChecker := awscrossplanerolechecker.New(c.logger, c.envConfig, iam.NewFromConfig(aws.Config{
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
		return nil, multierr.Combine(crossplanerolechecker.ErrFailedToCheckCrossplaneRole, err)
	}

	c.logger.Info(crossplanerolechecker.LogMsgCrossplaneRoleChecked)

	return nil, nil
}

// New is the function that creates a new AWSChecker.
func New(logger *log.Logger, envConfig *envconfig.EnvConfig, clientset kubernetes.Interface, httpClient *http.Client, jwksURI *string) *AWSChecker {
	c := &AWSChecker{
		logger:     logger,
		envConfig:  envConfig,
		clientset:  clientset,
		httpClient: httpClient,
		jwksURI:    jwksURI,
	}

	c.setup()

	return c
}
