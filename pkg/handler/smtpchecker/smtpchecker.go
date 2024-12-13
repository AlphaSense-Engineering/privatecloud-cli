// Package smtpchecker is the package that contains the check functions for the SMTP.
package smtpchecker

import (
	"context"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
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

		// secretAddressKey is the key of the address in the secret.
		secretAddressKey = "address"
		// secretHostKey is the key of the host in the secret.
		secretHostKey = "host"
	)

	secret, err := c.clientset.CoreV1().Secrets(constant.NamespaceAlphaSense).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := util.ConvertMap[string, []byte, string, string](secret.Data, util.Identity[string], util.ByteSliceToString)

	if err := util.KeysExistAndNotEmptyOrErr(data, []string{
		constant.SecretUsernameKey,
		constant.SecretPasswordKey,
		secretAddressKey,
		secretHostKey,
		constant.SecretPortKey,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// New is a function that returns a new SMTPChecker.
func New(clientset kubernetes.Interface) *SMTPChecker {
	return &SMTPChecker{clientset: clientset}
}
