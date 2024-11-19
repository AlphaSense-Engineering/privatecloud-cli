// Package smtpchecker is the package that contains the check functions for the SMTP.
package smtpchecker

import (
	"context"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SMTPChecker is the type that contains the check functions for the SMTP.
type SMTPChecker struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &SMTPChecker{}

// Handle is the function that handles the SMTP checking.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *SMTPChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// secretName is the name of the secret that contains the SMTP credentials.
		secretName = "sender-smtp" // nolint:gosec
	)

	_, err := c.clientset.CoreV1().Secrets(constant.NamespaceAlphaSense).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// TODO: Check the SMTP validity.

	return []any{}, nil
}

// New is a function that returns a new SMTP checker.
func New(clientset kubernetes.Interface) *SMTPChecker {
	return &SMTPChecker{clientset: clientset}
}
