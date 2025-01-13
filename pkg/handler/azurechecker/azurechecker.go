// Package azurechecker is the package that contains the check functions for Azure.
package azurechecker

import (
	"context"
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/azurecrossplanerolechecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/azurejwtretriever"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/crossplanerolechecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/jwtchecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/jwtretriever"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/charmbracelet/log"
	"go.uber.org/multierr"
	"k8s.io/client-go/kubernetes"
)

// AzureChecker is the type that contains the infrastructure check functions for Azure.
type AzureChecker struct {
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
	jwtRetriever *azurejwtretriever.AzureJWTRetriever
	// jwtChecker is the JWT checker.
	jwtChecker *jwtchecker.JWTChecker
}

var _ handler.Handler = &AzureChecker{}

// setup is the function that sets up the Azure checker.
func (c *AzureChecker) setup() {
	c.jwtRetriever = azurejwtretriever.New(c.clientset)

	c.jwtChecker = jwtchecker.New(c.httpClient, c.jwksURI)
}

// Handle is the function that handles the infrastructure check.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
//
// nolint:funlen
func (c *AzureChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	jwts, err := util.ConvertSliceErr[any, *string](c.jwtRetriever.Handle(ctx))
	if err != nil {
		return nil, multierr.Combine(jwtretriever.ErrFailedToRetrieveJWTs, err)
	}

	c.logger.Debug(jwtretriever.LogMsgJWTsRetrieved)

	if _, err := c.jwtChecker.Handle(ctx, jwts); err != nil {
		return nil, multierr.Combine(jwtchecker.ErrFailedToCheckJWTs, err)
	}

	c.logger.Debug(jwtchecker.LogMsgJWTsChecked)

	jwt := jwts[0]

	err = func() error {
		cred, err := azidentity.NewClientAssertionCredential(
			c.envConfig.Spec.CloudSpec.Azure.TenantID,
			c.envConfig.Spec.CloudSpec.Azure.ClientID,
			func(context.Context) (string, error) {
				return *jwt, nil
			},
			nil,
		)
		if err != nil {
			return err
		}

		roleDefClient, err := armauthorization.NewRoleDefinitionsClient(cred, nil)
		if err != nil {
			return err
		}

		crossplaneRoleChecker := azurecrossplanerolechecker.New(c.envConfig, roleDefClient)

		if _, err := crossplaneRoleChecker.Handle(ctx); err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		return nil, multierr.Combine(crossplanerolechecker.ErrFailedToCheckCrossplaneRole, err)
	}

	c.logger.Info(crossplanerolechecker.LogMsgCrossplaneRoleChecked)

	return nil, nil
}

// New is the function that creates a new AzureChecker.
func New(logger *log.Logger, envConfig *envconfig.EnvConfig, clientset kubernetes.Interface, httpClient *http.Client, jwksURI *string) *AzureChecker {
	c := &AzureChecker{
		logger:     logger,
		envConfig:  envConfig,
		clientset:  clientset,
		httpClient: httpClient,
		jwksURI:    jwksURI,
	}

	c.setup()

	return c
}
