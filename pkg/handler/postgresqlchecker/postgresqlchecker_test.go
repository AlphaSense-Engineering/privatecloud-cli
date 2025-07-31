// Package postgresqlchecker is the package that contains the check functions for the PostgreSQL.
package postgresqlchecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPostgreSQLChecker_buildConnString is a test that tests the buildConnString function.
func TestPostgreSQLChecker_buildConnString(t *testing.T) {
	c := &PostgreSQLChecker{}

	testCases := []struct {
		name     string
		username string
		password string
		endpoint string
		port     string
		want     string
	}{
		{
			name:     "Basic",
			username: "user",
			password: "pass",
			endpoint: "db.example.com",
			port:     "5432",
			want:     "postgresql://user:pass@db.example.com:5432/postgres?sslmode=disable",
		},
		{
			name:     "Special characters in password",
			username: "user",
			password: "p@ss:word",
			endpoint: "db.example.com",
			port:     "5432",
			want:     "postgresql://user:p%40ss%3Aword@db.example.com:5432/postgres?sslmode=disable",
		},
		{
			name:     "Empty password",
			username: "user",
			password: "",
			endpoint: "db.example.com",
			port:     "5432",
			want:     "postgresql://user:@db.example.com:5432/postgres?sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := c.buildConnString(tc.username, tc.password, tc.endpoint, tc.port)

			assert.Equal(t, tc.want, got, "expected %q, got %q", tc.want, got)
		})
	}
}
