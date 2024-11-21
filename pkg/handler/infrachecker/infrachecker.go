// Package infrachecker is the package that contains the infrastructure check functions.
package infrachecker

import (
	"context"
	"errors"
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/awschecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/azurechecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/cloudchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// errJWKSURIRequired is an error that occurs when the JWKS URI is required.
var errJWKSURIRequired = errors.New("jwks URI is required")

// New is the function that creates a new infrastructure checker.
func New(
	ctx context.Context,
	vcloud cloud.Cloud,
	envConfig *envconfig.EnvConfig,
	clientset kubernetes.Interface,
	httpClient *http.Client,
) (handler.Handler, error) {
	cloudChecker := cloudchecker.New(vcloud, envConfig, clientset, httpClient)

	jwksURI, err := util.FirstValErr(util.ConvertSliceErr[any, *string](cloudChecker.Handle(ctx)))
	if err != nil {
		return nil, err
	}

	if jwksURI == nil {
		return nil, errJWKSURIRequired
	}

	if vcloud == cloud.AWS {
		return awschecker.New(vcloud, envConfig, clientset, httpClient, jwksURI), nil
	} else if vcloud == cloud.Azure {
		return azurechecker.New(), nil
	}

	return nil, cloud.NewUnsupportedCloudError(vcloud)
}
