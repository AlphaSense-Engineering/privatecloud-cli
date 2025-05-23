// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	// errFailedToReadEnvConfig is the error that is returned when the environment configuration cannot be read.
	errFailedToReadEnvConfig = errors.New("failed to read environment configuration")

	// errFailedToGetKubeConfig is the error that is returned when the Kubernetes configuration cannot be retrieved.
	errFailedToGetKubeConfig = errors.New("failed to get Kubernetes configuration")

	// errFailedToCreateKubernetesClientset is the error that is returned when the Kubernetes clientset cannot be created.
	errFailedToCreateKubernetesClientset = errors.New("failed to create Kubernetes clientset")
)

const (
	// logMsgKubeLoadedConfig is the message that is logged when the Kubernetes configuration is loaded from the specified path.
	logMsgKubeLoadedConfig = "loaded Kubernetes configuration from %s"

	// logMsgKubeClientsetCreated is the message that is logged when the Kubernetes clientset is created.
	logMsgKubeClientsetCreated = "created Kubernetes clientset from configuration"
)

const (
	// envVarEnvConfig is the name of the environment variable that contains the base64 encoded environment configuration.
	envVarEnvConfig = "ENVCONFIG"

	// envVarGoogleCloudSDKDockerRepo is the name of the environment variable that contains the Docker repository for the Google Cloud SDK.
	envVarGoogleCloudSDKDockerRepo = "GOOGLE_CLOUD_SDK_DOCKER_REPO"

	// envVarGoogleCloudSDKDockerImage is the name of the environment variable that contains the Docker image for the Google Cloud SDK.
	envVarGoogleCloudSDKDockerImage = "GOOGLE_CLOUD_SDK_DOCKER_IMAGE"
)

// cmd is the interface that all commands must implement.
type cmd interface {
	// run is the run function for the command.
	run(*cobra.Command, []string)
}
