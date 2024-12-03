// Package gcpchecker is the package that contains the check functions for GCP.
package gcpchecker

import (
	"context"
	"log"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/crossplanerolechecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/gcpcrossplanerolechecker"
	"k8s.io/client-go/kubernetes"
)

// GCPChecker is the type that contains the infrastructure check functions for GCP.
type GCPChecker struct {
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
	c.crossplaneRoleChecker = gcpcrossplanerolechecker.New(c.envConfig, c.clientset)
}

// Handle is the function that handles the infrastructure check.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *GCPChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	if _, err := c.crossplaneRoleChecker.Handle(ctx); err != nil {
		return nil, crossplanerolechecker.ErrFailedToCheckCrossplaneRole
	}

	log.Println(crossplanerolechecker.LogMsgCrossplaneRoleChecked)

	return nil, nil
}

// New is the function that creates a new GCP checker.
func New(envConfig *envconfig.EnvConfig, clientset kubernetes.Interface) *GCPChecker {
	c := &GCPChecker{
		envConfig: envConfig,
		clientset: clientset,
	}

	c.setup()

	return c
}
