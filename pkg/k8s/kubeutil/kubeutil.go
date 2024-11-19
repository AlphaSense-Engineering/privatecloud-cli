// Package kubeutil provides utilities for interacting with Kubernetes.
package kubeutil

import (
	"os"
	"path/filepath"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config returns a Kubernetes configuration based on the provided path,
// or the path in the KUBECONFIG environment variable, or the default path.
func Config(path string) (config *rest.Config, pathToUse string, err error) {
	const (
		// kubeConfigEnvVar is the environment variable that contains the path to the Kubernetes configuration file.
		kubeConfigEnvVar = "KUBECONFIG"

		// pathHomeKubeDir is the Kubernetes directory name within the home directory.
		pathHomeKubeDir = ".kube"

		// pathKubeDirConfig is the Kubernetes configuration file name within the Kubernetes directory.
		pathKubeDirConfig = "config"
	)

	if path != constant.EmptyString {
		pathToUse = path
	} else if envPath := os.Getenv(kubeConfigEnvVar); envPath != constant.EmptyString {
		pathToUse = envPath
	} else {
		var pathHome string

		pathHome, err = os.UserHomeDir()
		if err != nil {
			return nil, constant.EmptyString, err
		}

		pathToUse = filepath.Join(pathHome, pathHomeKubeDir, pathKubeDirConfig)
	}

	if _, err = os.Stat(pathToUse); os.IsNotExist(err) {
		// pathToUseCluster is the path that we return when we are running in a cluster. This is not a real path,
		// it's just a placeholder to indicate that we are running in a cluster.
		const pathToUseCluster = "cluster"

		pathToUse = pathToUseCluster

		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, pathToUse, err
		}

		return config, pathToUse, nil
	}

	config, err = clientcmd.BuildConfigFromFlags(constant.EmptyString, pathToUse)
	if err != nil {
		return nil, pathToUse, err
	}

	return config, pathToUse, nil
}
