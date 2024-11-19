// Package mysqlutil is the package that contains the utility functions for MySQL.
package mysqlutil

import "fmt"

// DSN returns the Data Source Name (DSN) for the database.
func DSN(user string, pass string, host string, port string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, pass, host, port)
}
