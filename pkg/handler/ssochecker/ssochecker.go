// Package ssochecker is the package that contains the check functions for the SSO.
package ssochecker

import (
	"context"
	"errors"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// errSAMLEntityIDNotSet is the error that the SAML entity ID is not set.
	errSAMLEntityIDNotSet = errors.New("saml entity ID is not set")
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
		// secretNamespace is the namespace of the secret that contains the SSO configuration.
		secretNamespace = "platform"

		// secretName is the name of the secret that contains the SSO configuration.
		secretName = "sso-config" // nolint:gosec
	)

	secret, err := c.clientset.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := util.ConvertMap[string, []byte, string, string](secret.Data, util.Identity[string], util.ByteSliceToString)

	if samlEntityID, ok := data["saml-entityid"]; !ok || samlEntityID == constant.EmptyString {
		return nil, errSAMLEntityIDNotSet
	}

	return []any{}, nil
}

// New is a function that returns a new SSO checker.
func New(clientset kubernetes.Interface) *SSOChecker {
	return &SSOChecker{clientset: clientset}
}
