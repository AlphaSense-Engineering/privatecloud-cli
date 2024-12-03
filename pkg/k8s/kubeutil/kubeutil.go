// Package kubeutil provides utilities for interacting with Kubernetes.
package kubeutil

import (
	"bufio"
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// ErrFailedToGetPod is the error that is returned when the pod cannot be retrieved.
	ErrFailedToGetPod = errors.New("failed to get Pod")

	// errFailedToGetPodLogStream is the error that is returned when the pod log stream cannot be retrieved.
	errFailedToGetPodLogStream = errors.New("failed to get Pod log stream")

	// errFailedToReadPodLogStream is the error that is returned when the pod log stream cannot be read.
	errFailedToReadPodLogStream = errors.New("failed to read Pod log stream")
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

// WaitForPodToSucceedOrFail waits for the pod to succeed or fail.
func WaitForPodToSucceedOrFail(ctx context.Context, clientset kubernetes.Interface, namespace string, podName string) (phase corev1.PodPhase, err error) {
	const (
		// logMsgPodWaitingToSucceedOrFail is the message that is logged when we are waiting for the pod to succeed or fail.
		logMsgPodWaitingToSucceedOrFail = "waiting for %s/%s Pod to succeed or fail..."

		// logMsgPodSucceeded is the message that is logged when the pod succeeded.
		logMsgPodSucceeded = "%s/%s Pod succeeded"

		// logMsgPodFailed is the message that is logged when the pod failed.
		logMsgPodFailed = "%s/%s Pod failed"
	)

	log.Printf(logMsgPodWaitingToSucceedOrFail, namespace, podName)

	var pod *corev1.Pod

	for {
		pod, err = clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return corev1.PodUnknown, multierr.Combine(ErrFailedToGetPod, err)
		}

		phase = pod.Status.Phase

		if phase == corev1.PodSucceeded {
			log.Printf(logMsgPodSucceeded, namespace, podName)

			break
		} else if phase == corev1.PodFailed {
			log.Printf(logMsgPodFailed, namespace, podName)

			break
		}

		time.Sleep(time.Second)
	}

	return phase, nil
}

// PodLogs retrieves the pod logs.
func PodLogs(ctx context.Context, clientset kubernetes.Interface, namespace string, podName string) ([]string, error) {
	// logMsgPodLogStreamRetrieved is the message that is logged when the pod log stream is retrieved.
	const logMsgPodLogStreamRetrieved = "retrieved log stream for %s/%s Pod, printing..."

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{})

	podLogStream, err := req.Stream(ctx)
	if err != nil {
		return nil, multierr.Combine(errFailedToGetPodLogStream, err)
	}
	defer podLogStream.Close() // nolint:errcheck

	log.Printf(logMsgPodLogStreamRetrieved, namespace, podName)

	var logLines []string

	scanner := bufio.NewScanner(podLogStream)

	for scanner.Scan() {
		trimmedLine := strings.TrimSpace(scanner.Text())

		if trimmedLine != constant.EmptyString {
			logLines = append(logLines, trimmedLine)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, multierr.Combine(errFailedToReadPodLogStream, err)
	}

	return logLines, nil
}
