// Package azurechecker is the package that contains the check functions for Azure.
package azurechecker

import (
	"context"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
)

// AzureChecker is the type that contains the infrastructure check functions for Azure.
type AzureChecker struct{}

var _ handler.Handler = &AzureChecker{}

// setup is the function that sets up the Azure checker.
func (c *AzureChecker) setup() {}

// Handle is the function that handles the infrastructure check.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
//
// nolint:funlen
func (c *AzureChecker) Handle(_ context.Context, _ ...any) ([]any, error) {
	return []any{}, nil
}

// New is the function that creates a new Azure checker.
func New() *AzureChecker {
	c := &AzureChecker{}

	c.setup()

	return c
}
