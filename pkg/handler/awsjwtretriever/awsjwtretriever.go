// Package awsjwtretriever contains the JWT retriever for AWS.
package awsjwtretriever

import (
	"context"
	"strings"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/jwtretriever"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// AWSJWTRetriever is the JWT retriever for AWS.
type AWSJWTRetriever struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &AWSJWTRetriever{}

// Handle is the function that handles the JWT retrieval for AWS.
//
// The arguments are not used.
// It returns a slice of JWTs on success, or an error on failure.
func (c *AWSJWTRetriever) Handle(ctx context.Context, _ ...any) (jwts []any, err error) {
	const (
		// serviceAccountsPrefix is the prefix of the service accounts in AWS configuration.
		serviceAccountsPrefix = "aws-"

		// audience is the audience of the AWS JWTs.
		audience = "amazonaws.com"
	)

	clientsetSA := c.clientset.CoreV1().ServiceAccounts(constant.NamespaceCrossplane)

	serviceAccounts, err := clientsetSA.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, sa := range serviceAccounts.Items {
		if !strings.HasPrefix(sa.Name, serviceAccountsPrefix) {
			continue
		}

		req, err := clientsetSA.CreateToken(ctx, sa.Name, &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{audience},
				ExpirationSeconds: util.Ref(jwtretriever.TokenExpirationSeconds),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}

		if req.Status.Token != constant.EmptyString {
			jwts = append(jwts, &req.Status.Token)
		}
	}

	if jwts == nil {
		err = jwtretriever.ErrNoJWTsRetrieved
	}

	return jwts, err
}

// New creates a new AWSJWTRetriever.
func New(clientset kubernetes.Interface) *AWSJWTRetriever {
	return &AWSJWTRetriever{clientset: clientset}
}
