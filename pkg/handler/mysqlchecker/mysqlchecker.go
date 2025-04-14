// Package mysqlchecker is the package that contains the check functions for the MySQL.
package mysqlchecker

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
	pkgerrors "github.com/AlphaSense-Engineering/privatecloud-cli/pkg/errors"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	"github.com/go-sql-driver/mysql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// constExpectedConfig is the map of expected configuration for the MySQL.
	//
	// These are listed at https://developer.alpha-sense.com/enterprise/technical-requirements/#mysql-database-cluster.
	//
	// Do not modify this variable, it is supposed to be constant.
	constExpectedConfig = map[string]string{
		"connect_timeout":                 "20",
		"explicit_defaults_for_timestamp": "1",
		"innodb_print_all_deadlocks":      "1",
		"lower_case_table_names":          "1",
		"net_read_timeout":                "60",
		"net_write_timeout":               "120",
		"require_secure_transport":        "0",
		"wait_timeout":                    "1800",
	}
)

// MySQLChecker is the type that contains the check functions for the MySQL.
type MySQLChecker struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &MySQLChecker{}

// Handle is the function that handles the MySQL checking.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *MySQLChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// secretName is the name of the secret that contains the MySQL credentials.
		secretName = "default-creds"

		// secretEndpointKey is the key of the endpoint in the secret.
		secretEndpointKey = "endpoint"
	)

	secret, err := c.clientset.CoreV1().Secrets(constant.NamespaceMySQL).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := util.ConvertMap[string, []byte, string, string](secret.Data, util.Identity[string], util.ByteSliceToString)

	if err := util.KeysExistAndNotEmptyOrErr(data, []string{
		constant.SecretUsernameKey,
		constant.SecretPasswordKey,
		secretEndpointKey,
		constant.SecretPortKey,
	}); err != nil {
		return nil, err
	}

	cfg := mysql.NewConfig()

	cfg.User = data[constant.SecretUsernameKey]
	cfg.Passwd = data[constant.SecretPasswordKey]
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%s", data[secretEndpointKey], data[constant.SecretPortKey])

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	defer db.Close() // nolint:errcheck

	if err := db.Ping(); err != nil {
		return nil, err
	}

	for k, expected := range constExpectedConfig {
		var got string
		if err := db.QueryRowContext(ctx, "SELECT @@"+k).Scan(&got); err != nil {
			return nil, err
		}

		if got != expected {
			return nil, pkgerrors.NewKeyExpectedGot(k, expected, got)
		}
	}

	return nil, nil
}

// New is a function that returns a new MySQLChecker.
func New(clientset kubernetes.Interface) *MySQLChecker {
	return &MySQLChecker{clientset: clientset}
}
