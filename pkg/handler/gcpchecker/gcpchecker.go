// Package gcpchecker is the package that contains the check functions for GCP.
package gcpchecker

import (
	"context"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/crossplanerolechecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/gcpcrossplanerolechecker"
	"github.com/charmbracelet/log"
	"k8s.io/client-go/kubernetes"
)

// GCPChecker is the type that contains the infrastructure check functions for GCP.
type GCPChecker struct {
	// logger is the logger.
	logger *log.Logger
	// envConfig is the environment configuration.
	envConfig *envconfig.EnvConfig
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface

	// crossplaneRoleChecker is the GCP Crossplane role checker.
	crossplaneRoleChecker *gcpcrossplanerolechecker.GCPCrossplaneRoleChecker
}

var _ handler.Handler = &GCPChecker{}

// setup is the function that sets up the GCP checker.
func (c *GCPChecker) setup() {
	c.crossplaneRoleChecker = gcpcrossplanerolechecker.New(c.logger, c.envConfig, c.clientset)
}

// Handle is the function that handles the infrastructure check.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *GCPChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	if _, err := c.crossplaneRoleChecker.Handle(ctx); err != nil {
		return nil, crossplanerolechecker.ErrFailedToCheckCrossplaneRole
	}

	c.logger.Info(crossplanerolechecker.LogMsgCrossplaneRoleChecked)

	return nil, nil
}

// New is the function that creates a new GCPChecker.
func New(logger *log.Logger, envConfig *envconfig.EnvConfig, clientset kubernetes.Interface) *GCPChecker {
	c := &GCPChecker{
		logger:    logger,
		envConfig: envConfig,
		clientset: clientset,
	}

	c.setup()

	return c
}
