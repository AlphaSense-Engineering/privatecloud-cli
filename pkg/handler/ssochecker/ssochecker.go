// Package ssochecker is the package that contains the check functions for the SSO.
package ssochecker

import (
	"context"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SSOChecker is the type that contains the check functions for the SSO.
type SSOChecker struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &SSOChecker{}

// Handle is the function that handles the SSO checking.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *SSOChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// secretName is the name of the secret that contains the SSO configuration.
		secretName = "sso-config" // nolint:gosec
	)

	secret, err := c.clientset.CoreV1().Secrets(constant.NamespacePlatform).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := util.ConvertMap[string, []byte, string, string](secret.Data, util.Identity[string], util.ByteSliceToString)

	if err := util.KeysExistAndNotEmptyOrErr(data, []string{"saml-entityid"}); err != nil {
		return nil, err
	}

	return nil, nil
}

// New is a function that returns a new SSOChecker.
func New(clientset kubernetes.Interface) *SSOChecker {
	return &SSOChecker{clientset: clientset}
}
