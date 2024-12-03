// Package azurejwtretriever contains the JWT retriever for Azure.
package azurejwtretriever

import (
	"context"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/jwtretriever"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// AzureJWTRetriever is the JWT retriever for Azure.
type AzureJWTRetriever struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &AzureJWTRetriever{}

// Handle is the function that handles the JWT retrieval for Azure.
//
// The arguments are not used.
// It returns a slice of JWTs on success, or an error on failure.
func (c *AzureJWTRetriever) Handle(ctx context.Context, _ ...any) (jwts []any, err error) {
	const (
		// audience is the audience of the Azure JWTs.
		audience = "api://AzureADTokenExchange"
	)

	clientsetSA := c.clientset.CoreV1().ServiceAccounts(constant.NamespaceCrossplane)

	req, err := clientsetSA.CreateToken(ctx, constant.ServiceAccountNameAzure, &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences:         []string{audience},
			ExpirationSeconds: util.Ref(jwtretriever.TokenExpirationSeconds),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return
	}

	if req.Status.Token != constant.EmptyString {
		jwts = append(jwts, &req.Status.Token)
	}

	if jwts == nil {
		err = jwtretriever.ErrNoJWTsRetrieved
	}

	return jwts, err
}

// New creates a new AzureJWTRetriever.
func New(clientset kubernetes.Interface) *AzureJWTRetriever {
	return &AzureJWTRetriever{clientset: clientset}
}
