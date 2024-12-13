// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"os"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud/gcpcloudutil"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	pkgerrors "github.com/AlphaSense-Engineering/privatecloud-installer/pkg/errors"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/awschecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/azurechecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/cloudchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler/gcpchecker"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/k8s/kubeutil"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// errEnvConfigFlagIsNotSetOrEmpty is the error that is returned when the envconfig flag is not set or empty.
	errEnvConfigFlagIsNotSetOrEmpty = errors.New("envconfig flag is not set or empty")

	// errFailedToDecodeEnvConfig is the error that is returned when the envconfig data from the flag cannot be decoded.
	errFailedToDecodeEnvConfig = errors.New("failed to decode envconfig")

	// errFailedToEnsureNamespace is the error that is returned when the namespace cannot be ensured.
	errFailedToEnsureNamespace = errors.New("failed to ensure Namespace")

	// errFailedToEnsureServiceAccount is the error that is returned when the service account cannot be ensured.
	errFailedToEnsureServiceAccount = errors.New("failed to ensure ServiceAccount")

	// errJWKSURIRequired is an error that occurs when the JWKS URI is required.
	errJWKSURIRequired = errors.New("jwks URI is required")

	// errFailedToCheckInfrastructure is the error that is returned when the infrastructure check fails.
	errFailedToCheckInfrastructure = errors.New("failed to check infrastructure")
)

// podCmd is the command that checks the infrastructure of the cluster where it is running on.
type podCmd struct {
	// logger is the logger.
	logger *log.Logger
}

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

		// logMsgNamespaceEnsured is the message that is logged when the namespace is ensured.
		logMsgNamespaceEnsured = "ensured %s Namespace"

		// logMsgServiceAccountEnsured is the message that is logged when the service account is ensured.
		logMsgServiceAccountEnsured = "ensured %s/%s ServiceAccount"

		// logMsgInfraCheckCompletedSuccessfully is the message that is logged when the infrastructure check is completed successfully.
		logMsgInfraCheckCompletedSuccessfully = "infrastructure check completed successfully"
	)

	c.logger.SetFormatter(log.JSONFormatter)

	c.logger.Logf(log.InfoLevel, logMsgPodStarted, constant.AppName)

	envConfigBase64 := os.Getenv(envVarEnvConfig)
	if envConfigBase64 == constant.EmptyString {
		c.logger.Fatal(errEnvConfigFlagIsNotSetOrEmpty)
	}

	envConfigBytes, err := base64.StdEncoding.DecodeString(envConfigBase64)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToDecodeEnvConfig, err))
	}

	envConfig, err := envconfig.NewFromBytes(envConfigBytes)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToReadEnvConfig, err))
	}

	c.logger.Log(log.InfoLevel, logMsgEnvConfigDecoded)

	kubeConfig, path, err := kubeutil.Config(constant.EmptyString)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToGetKubeConfig, err))
	}

	c.logger.Logf(log.InfoLevel, logMsgKubeLoadedConfig, path)

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToCreateKubernetesClientset, err))
	}

	c.logger.Log(log.InfoLevel, logMsgKubeClientsetCreated)

	vcloud := cloud.Cloud(envConfig.Spec.CloudSpec.Provider)

	ctx := context.Background()

	if _, err := clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.NamespaceCrossplane,
		},
	}, metav1.CreateOptions{}); err != nil && !k8serrors.IsAlreadyExists(err) {
		c.logger.Fatal(multierr.Combine(errFailedToEnsureNamespace, err))
	}

	c.logger.Logf(log.InfoLevel, logMsgNamespaceEnsured, constant.NamespaceCrossplane)

	var serviceAccountName string

	if vcloud == cloud.AWS {
		serviceAccountName = constant.ServiceAccountNameAWS
	} else if vcloud == cloud.Azure {
		serviceAccountName = constant.ServiceAccountNameAzure
	} else if vcloud == cloud.GCP {
		serviceAccountName = constant.ServiceAccountNameGCP
	} else {
		c.logger.Fatal(pkgerrors.NewUnsupportedCloud(vcloud))
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: constant.NamespaceCrossplane,
		},
	}

	if vcloud == cloud.GCP {
		sa.ObjectMeta.Annotations = map[string]string{
			"iam.gke.io/gcp-service-account": gcpcloudutil.ServiceAccountAnnotation(envConfig.Spec.ClusterName, envConfig.Spec.CloudSpec.GCP.ProjectID),
		}
	}

	if _, err = clientset.CoreV1().ServiceAccounts(constant.NamespaceCrossplane).Create(
		ctx, sa, metav1.CreateOptions{},
	); err != nil && !k8serrors.IsAlreadyExists(err) {
		c.logger.Fatal(multierr.Combine(errFailedToEnsureServiceAccount, err))
	}

	c.logger.Logf(log.InfoLevel, logMsgServiceAccountEnsured, constant.NamespaceCrossplane, serviceAccountName)

	httpClient := http.DefaultClient

	cloudChecker := cloudchecker.New(c.logger, vcloud, envConfig, clientset, httpClient)

	var jwksURI *string

	rawJWKSURI, err := cloudChecker.Handle(ctx)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToCheckInfrastructure, err))
	}

	if rawJWKSURI != nil {
		jwksURI, _ = rawJWKSURI[0].(*string)
	}

	// In GCP, we don't need to check the OIDC URL as it's not used.
	if vcloud != cloud.GCP && jwksURI == nil {
		c.logger.Fatal(multierr.Combine(errFailedToCheckInfrastructure, errJWKSURIRequired))
	}

	var concreteCloudChecker handler.Handler

	if vcloud == cloud.AWS {
		concreteCloudChecker = awschecker.New(c.logger, envConfig, clientset, httpClient, jwksURI)
	} else if vcloud == cloud.Azure {
		concreteCloudChecker = azurechecker.New(c.logger, envConfig, clientset, httpClient, jwksURI)
	} else if vcloud == cloud.GCP {
		concreteCloudChecker = gcpchecker.New(c.logger, envConfig, clientset)
	}

	if _, err := concreteCloudChecker.Handle(ctx); err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToCheckInfrastructure, err))
	}

	c.logger.Log(log.InfoLevel, logMsgInfraCheckCompletedSuccessfully)
}

// newPodCmd returns a new podCmd.
func newPodCmd(logger *log.Logger) *podCmd {
	return &podCmd{
		logger: logger,
	}
}

// Pod returns a Cobra command that checks the infrastructure of the cluster where it is running on.
func Pod(logger *log.Logger) *cobra.Command {
	cmd := newPodCmd(logger)

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
