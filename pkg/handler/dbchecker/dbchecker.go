// Package dbchecker is the package that contains the check functions for the database.
package dbchecker

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/db/mysqlutil"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// configMismatchError is the error that is returned when the configuration of the database is mismatched.
type configMismatchError struct {
	// key is the key that is mismatched.
	key string
	// expected is the expected value.
	expected string
	// got is the got value.
	got string
}

var _ error = &configMismatchError{}

// Error is a function that returns the error message.
func (e *configMismatchError) Error() string {
	return fmt.Sprintf("expected %s to be %s, got %s", e.key, e.expected, e.got)
}

// newConfigMismatchError is a function that returns a new config mismatch error.
func newConfigMismatchError(key, expected, got string) *configMismatchError {
	return &configMismatchError{key: key, expected: expected, got: got}
}

var (
	// constExpectedConfig is the map of expected configuration for the database.
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

// DBChecker is the type that contains the check functions for the database.
type DBChecker struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &DBChecker{}

// Handle is the function that handles the database checking.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *DBChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// secretName is the name of the secret that contains the database credentials.
		secretName = "default-creds"
	)

	secret, err := c.clientset.CoreV1().Secrets(constant.NamespaceMySQL).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := util.ConvertMap[string, []byte, string, string](secret.Data, util.Identity[string], util.ByteSliceToString)

	dsn := mysqlutil.DSN(data["username"], data["password"], data["endpoint"], data["port"])

	db, err := sql.Open("mysql", dsn)
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
			return nil, newConfigMismatchError(k, expected, got)
		}
	}

	return []any{}, nil
}

// New is a function that returns a new database checker.
func New(clientset kubernetes.Interface) *DBChecker {
	return &DBChecker{clientset: clientset}
}
