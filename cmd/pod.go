// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/infrachecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/k8s/kubeutil"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// errEnvConfigFlagIsNotSetOrEmpty is the error that is returned when the envconfig flag is not set or empty.
	errEnvConfigFlagIsNotSetOrEmpty = errors.New("envconfig flag is not set or empty")

	// errFailedToDecodeEnvConfig is the error that is returned when the envconfig data from the flag cannot be decoded.
	errFailedToDecodeEnvConfig = errors.New("failed to decode envconfig")

	// errFailedToCreateInfraChecker is the error that is returned when the infrastructure checker cannot be created.
	errFailedToCreateInfraChecker = errors.New("failed to create infrastructure checker")

	// errFailedToCheckInfrastructure is the error that is returned when the infrastructure check fails.
	errFailedToCheckInfrastructure = errors.New("failed to check infrastructure")
)

// podCmd is the command that checks the infrastructure of the cluster where it is running on.
type podCmd struct{}

var _ cmd = &podCmd{}

// Run is the run function for the Pod command.
//
// nolint:funlen
func (c *podCmd) Run(_ *cobra.Command, _ []string) {
	const (
		// logMsgPodStarted is the message that is logged when the pod starts.
		logMsgPodStarted = "pod %s started"

		// logMsgEnvConfigDecoded is the message that is logged when the environment configuration is decoded.
		logMsgEnvConfigDecoded = "decoded environment configuration"

		// logMsgInfraCheckCompletedSuccessfully is the message that is logged when the infrastructure check is completed successfully.
		logMsgInfraCheckCompletedSuccessfully = "infrastructure check completed successfully"
	)

	log.SetFlags(constant.LogFlags)

	log.Printf(logMsgPodStarted, constant.AppName)

	envConfigBase64 := os.Getenv(envVarEnvConfig)
	if envConfigBase64 == constant.EmptyString {
		log.Fatalln(errEnvConfigFlagIsNotSetOrEmpty)

		return
	}

	envConfigBytes, err := base64.StdEncoding.DecodeString(envConfigBase64)
	if err != nil {
		log.Fatalln(multierr.Combine(errFailedToDecodeEnvConfig, err))

		return
	}

	envConfig, err := envconfig.NewFromBytes(envConfigBytes)
	if err != nil {
		log.Fatalln(multierr.Combine(errFailedToReadEnvConfig, err))

		return
	}

	log.Println(logMsgEnvConfigDecoded)

	kubeConfig, path, err := kubeutil.Config(constant.EmptyString)
	if err != nil {
		log.Fatalln(multierr.Combine(errFailedToGetKubeConfig, err))
	}

	log.Printf(logMsgKubeLoadedConfig, path)

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalln(multierr.Combine(errFailedToCreateKubernetesClientset, err))
	}

	log.Println(logMsgKubeClientsetCreated)

	vcloud := cloud.Cloud(envConfig.Spec.CloudSpec.Provider)

	const (
		// serviceAccountNameAWS is the name of the service account for AWS.
		serviceAccountNameAWS = "aws-privatecloud-installer"

		// serviceAccountNameAzure is the name of the service account for Azure.
		serviceAccountNameAzure = "azure-provider-sa"

		// serviceAccountNameGCP is the name of the service account for GCP.
		serviceAccountNameGCP = "gcp-provider-sa"
	)

	var serviceAccountName string

	if vcloud == cloud.AWS {
		serviceAccountName = serviceAccountNameAWS
	} else if vcloud == cloud.Azure {
		serviceAccountName = serviceAccountNameAzure
	} else if vcloud == cloud.GCP {
		serviceAccountName = serviceAccountNameGCP
	} else {
		log.Fatalln(cloud.NewUnsupportedCloudError(vcloud))
	}

	ctx := context.Background()

	_, err = clientset.CoreV1().ServiceAccounts(namespaceCrossplane).Create(ctx,
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceAccountName,
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		log.Fatalln(multierr.Combine(errFailedToCreateServiceAccount, err))
	}

	log.Printf(logMsgServiceAccountCreated, namespaceCrossplane, serviceAccountName)

	infraChecker, err := infrachecker.New(ctx, vcloud, envConfig, clientset, http.DefaultClient)
	if err != nil {
		log.Fatalln(multierr.Combine(errFailedToCreateInfraChecker, err))

		return
	}

	if _, err := infraChecker.Handle(ctx); err != nil {
		log.Fatalln(multierr.Combine(errFailedToCheckInfrastructure, err))

		return
	}

	log.Println(logMsgInfraCheckCompletedSuccessfully)
}

// newPodCmd returns a new Pod command.
func newPodCmd() *podCmd {
	return &podCmd{}
}

// Pod returns a Cobra command that checks the infrastructure of the cluster where it is running on.
func Pod() *cobra.Command {
	cmd := newPodCmd()

	return &cobra.Command{
		Use:   "pod",
		Short: "Run the pod",
		Long: `Pod checks the infrastructure of the cluster where it is running on before the installer can proceed.

When running this command, provide the environment configuration as a base64 encoded YAML file via the ENVCONFIG environment variable.

You are not supposed to run this command manually, unless you know what you are doing.`,
		Run:    cmd.Run,
		Hidden: true,
	}
}
