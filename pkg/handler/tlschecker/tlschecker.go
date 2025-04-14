// Package tlschecker is the package that contains the check functions for the TLS.
package tlschecker

import (
	"context"
	"crypto/tls"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	corev1 "k8s.io/api/core/v1"
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
// It returns the TLS secret on success, or an error on failure.
func (c *TLSChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// secretName is the name of the secret that contains the TLS credentials.
		secretName = "default-tls"
	)

	secret, err := c.clientset.CoreV1().Secrets(constant.NamespaceAlphaSense).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := secret.Data

	if err := util.KeysExistAndNotEmptyOrErr(data, []string{corev1.TLSCertKey, corev1.TLSPrivateKeyKey}); err != nil {
		return nil, err
	}

	if _, err = tls.X509KeyPair(data[corev1.TLSCertKey], data[corev1.TLSPrivateKeyKey]); err != nil {
		return nil, err
	}

	return []any{secret}, nil
}

// New is a function that returns a new TLSChecker.
func New(clientset kubernetes.Interface) *TLSChecker {
	return &TLSChecker{clientset: clientset}
}
