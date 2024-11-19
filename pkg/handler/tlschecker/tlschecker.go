// Package tlschecker is the package that contains the check functions for the TLS.
package tlschecker

import (
	"context"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TLSChecker is the type that contains the check functions for the TLS.
type TLSChecker struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &TLSChecker{}

// Handle is the function that handles the TLS checking.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *TLSChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// secretName is the name of the secret that contains the TLS credentials.
		secretName = "default-tls"
	)

	_, err := c.clientset.CoreV1().Secrets(constant.NamespaceAlphaSense).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// TODO: Check the TLS validity.

	return []any{}, nil
}

// New is a function that returns a new TLS checker.
func New(clientset kubernetes.Interface) *TLSChecker {
	return &TLSChecker{clientset: clientset}
}
