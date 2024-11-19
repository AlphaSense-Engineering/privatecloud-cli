// Package cmd is the package that contains all of the commands for the application.
package cmd

const (
	// logMsgKubeLoadedConfig is the message that is logged when the Kubernetes configuration is loaded from the specified path.
	logMsgKubeLoadedConfig = "loaded Kubernetes configuration from %s"

	// logMsgKubeClientsetCreated is the message that is logged when the Kubernetes clientset is created.
	logMsgKubeClientsetCreated = "created Kubernetes clientset from configuration"

	// logMsgServiceAccountCreated is the message that is logged when the service account is created.
	logMsgServiceAccountCreated = "created %s/%s ServiceAccount"
)

const (
	// envVarEnvConfig is the name of the environment variable that contains the base64 encoded environment configuration.
	envVarEnvConfig = "ENVCONFIG"
)

const (
	// namespaceCrossplane is the namespace for the Crossplane.
	namespaceCrossplane = "crossplane"
)
