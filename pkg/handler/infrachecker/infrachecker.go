// Package infrachecker is the package that contains the infrastructure check functions.
package infrachecker

import (
	"net/http"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/awschecker"
	"k8s.io/client-go/kubernetes"
)

// New is the function that creates a new infrastructure checker.
func New(
	vcloud cloud.Cloud,
	envConfig *envconfig.EnvConfig,
	clientset kubernetes.Interface,
	httpClient *http.Client,
) (handler.Handler, error) {
	if vcloud == cloud.AWS {
		return awschecker.New(vcloud, envConfig, clientset, httpClient), nil
	}

	return nil, cloud.NewUnsupportedCloudError(vcloud)
}
