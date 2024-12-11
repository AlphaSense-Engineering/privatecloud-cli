// Package smtpchecker is the package that contains the check functions for the SMTP.
package smtpchecker

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
	// errUsernameNotSet is the error that the username is not set.
	errUsernameNotSet = errors.New("username is not set")

	// errPasswordNotSet is the error that the password is not set.
	errPasswordNotSet = errors.New("password is not set")

	// errAddressNotSet is the error that the address is not set.
	errAddressNotSet = errors.New("address is not set")

	// errHostNotSet is the error that the host is not set.
	errHostNotSet = errors.New("host is not set")

	// errPortNotSet is the error that the port is not set.
	errPortNotSet = errors.New("port is not set")
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

	secret, err := c.clientset.CoreV1().Secrets(constant.NamespaceAlphaSense).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := util.ConvertMap[string, []byte, string, string](secret.Data, util.Identity[string], util.ByteSliceToString)

	if username, ok := data["username"]; !ok || username == constant.EmptyString {
		return nil, errUsernameNotSet
	}

	if password, ok := data["password"]; !ok || password == constant.EmptyString {
		return nil, errPasswordNotSet
	}

	if address, ok := data["address"]; !ok || address == constant.EmptyString {
		return nil, errAddressNotSet
	}

	if host, ok := data["host"]; !ok || host == constant.EmptyString {
		return nil, errHostNotSet
	}

	if port, ok := data["port"]; !ok || port == constant.EmptyString {
		return nil, errPortNotSet
	}

	return nil, nil
}

// New is a function that returns a new SMTPChecker.
func New(clientset kubernetes.Interface) *SMTPChecker {
	return &SMTPChecker{clientset: clientset}
}
