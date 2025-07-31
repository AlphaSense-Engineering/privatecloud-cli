// Package postgresqlchecker is the package that contains the check functions for the PostgreSQL.
package postgresqlchecker

import (
	"context"
	"fmt"
	"net/url"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	"github.com/jackc/pgx/v5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PostgreSQLChecker is the type that contains the check functions for the PostgreSQL.
type PostgreSQLChecker struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &PostgreSQLChecker{}

// buildConnString is a function that builds the connection string for the PostgreSQL.
func (c *PostgreSQLChecker) buildConnString(username string, password string, endpoint string, port string) string {
	const (
		// scheme is the scheme of the connection string.
		scheme = "postgresql"

		// database is the name of the database to connect to.
		database = "postgres"

		// sslmodeKey is the key of the SSL mode in the query parameters.
		sslmodeKey = "sslmode"

		// sslmodeDisable is the disabled SSL mode value in the query parameters.
		sslmodeDisable = "disable"
	)

	u := &url.URL{
		Scheme: scheme,
		User:   url.UserPassword(username, password),
		Host:   fmt.Sprintf("%s:%s", endpoint, port),
		Path:   database,
		RawQuery: url.Values{
			sslmodeKey: []string{sslmodeDisable},
		}.Encode(),
	}

	return u.String()
}

// Handle is the function that handles the PostgreSQL checking.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *PostgreSQLChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// secretName is the name of the secret that contains the PostgreSQL credentials.
		//
		// nolint:gosec
		secretName = "spicedb-creds"
	)

	secret, err := c.clientset.CoreV1().Secrets(constant.NamespacePostgres).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := util.ConvertMap(secret.Data, util.Identity[string], util.ByteSliceToString)

	if err := util.KeysExistAndNotEmptyOrErr(data, []string{
		constant.SecretUsernameKey,
		constant.SecretPasswordKey,
		constant.SecretEndpointKey,
		constant.SecretPortKey,
	}); err != nil {
		return nil, err
	}

	connString := c.buildConnString(
		data[constant.SecretUsernameKey],
		data[constant.SecretPasswordKey],
		data[constant.SecretEndpointKey],
		data[constant.SecretPortKey],
	)

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx) // nolint:errcheck

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return nil, nil
}

// New is a function that returns a new PostgreSQLChecker.
func New(clientset kubernetes.Interface) *PostgreSQLChecker {
	return &PostgreSQLChecker{clientset: clientset}
}
