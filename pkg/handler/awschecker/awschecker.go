// Package awschecker is the package that contains the check functions for AWS.
package awschecker

import (
	"context"
	"log"
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud/awscloudutil"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/awscrossplanerolechecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/awsjwtretriever"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/cloudchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/jwtchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.uber.org/multierr"
	"k8s.io/client-go/kubernetes"
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
		return nil, multierr.Combine(cloudchecker.ErrFailedToRetrieveJWTs, err)
	}

	log.Println(cloudchecker.LogMsgJWTsRetrieved)

	if _, err := c.jwtChecker.Handle(ctx, jwts); err != nil {
		return nil, multierr.Combine(cloudchecker.ErrFailedToCheckJWTs, err)
	}

	log.Println(cloudchecker.LogMsgJWTsChecked)

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
		return nil, multierr.Combine(cloudchecker.ErrFailedToCheckCrossplaneRole, err)
	}

	log.Println(cloudchecker.LogMsgCrossplaneRoleChecked)

	return []any{}, nil
}

// New is the function that creates a new AWS checker.
func New(vcloud cloud.Cloud, envConfig *envconfig.EnvConfig, clientset kubernetes.Interface, httpClient *http.Client, jwksURI *string) *AWSChecker {
	c := &AWSChecker{
		vcloud:     vcloud,
		envConfig:  envConfig,
		clientset:  clientset,
		httpClient: httpClient,
		jwksURI:    jwksURI,
	}

	c.setup()

	return c
}
