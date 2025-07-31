// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"os"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/cloud"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/cloud/gcpcloudutil"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/envconfig"
	pkgerrors "github.com/AlphaSense-Engineering/privatecloud-cli/pkg/errors"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/awschecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/azurechecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/cloudchecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler/gcpchecker"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/k8s/kubeutil"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// errFailedToDecodeEnvConfig is the error that is returned when the envconfig data from the flag cannot be decoded.
	errFailedToDecodeEnvConfig = errors.New("failed to decode envconfig")

	// errFailedToEnsureServiceAccount is the error that is returned when the service account cannot be ensured.
	errFailedToEnsureServiceAccount = errors.New("failed to ensure ServiceAccount")

	// errJWKSURIRequired is an error that occurs when the JWKS URI is required.
	errJWKSURIRequired = errors.New("jwks URI is required")

	// errUnknownError is the error that is returned when the error is unknown.
	errUnknownError = errors.New("unknown error")

	// errFailedToCheckInfrastructure is the error that is returned when the infrastructure check fails.
	errFailedToCheckInfrastructure = errors.New("failed to check infrastructure")
)

// podCmd is the command that checks the infrastructure of the cluster where it is running on.
type podCmd struct {
	// logger is the logger.
	logger *log.Logger
}

var _ cmd = &podCmd{}

func (c *podCmd) logRelatedDocumentation(docs ...string) {
	const (
		// logMsgRelatedDocumentation is the message that is logged when the related documentation resources are logged.
		logMsgRelatedDocumentation = "related documentation resources:"

		// logMsgRelatedDocumentationListPrefix is the prefix that is used when the list of documentation resources is logged.
		logMsgRelatedDocumentationListPrefix = "  - "
	)

	c.logger.Info(logMsgRelatedDocumentation)

	for _, doc := range docs {
		c.logger.Infof("%s%s", logMsgRelatedDocumentationListPrefix, doc)
	}
}

// run is the run function for the Pod command.
//
// nolint:funlen,gocognit
func (c *podCmd) run(_ *cobra.Command, _ []string) {
	const (
		// logMsgPodStarted is the message that is logged when the pod starts.
		logMsgPodStarted = "pod %s started"

		// logMsgEnvConfigDecoded is the message that is logged when the environment configuration is decoded.
		logMsgEnvConfigDecoded = "decoded environment configuration"

		// logMsgServiceAccountEnsured is the message that is logged when the service account is ensured.
		logMsgServiceAccountEnsured = "ensured %s/%s ServiceAccount"

		// logMsgInfraCheckCompletedSuccessfully is the message that is logged when the infrastructure check is completed successfully.
		logMsgInfraCheckCompletedSuccessfully = "infrastructure check completed successfully"
	)

	const (
		// docsPersistentVolumes is the URL to the documentation for persistent volumes.
		docsPersistentVolumes = "https://developer.alpha-sense.com/enterprise/technical-requirements/#persistent-volumes"

		// docsMySQLDatabaseCluster is the URL to the documentation for MySQL database cluster.
		docsMySQLDatabaseCluster = "https://developer.alpha-sense.com/enterprise/technical-requirements/#mysql-database-cluster"

		// docsMySQLSecrets is the URL to the documentation for MySQL secrets.
		//
		// nolint:gosec
		docsMySQLSecrets = "https://developer.alpha-sense.com/enterprise/technical-requirements/#mysql-secrets"

		// docsPostgreSQLDatabaseCluster is the URL to the documentation for PostgreSQL database cluster.
		docsPostgreSQLDatabaseCluster = "https://developer.alpha-sense.com/enterprise/technical-requirements/#postgresql-database-cluster"

		// docsPostgreSQLSecrets is the URL to the documentation for PostgreSQL secrets.
		//
		// nolint:gosec
		docsPostgreSQLSecrets = "https://developer.alpha-sense.com/enterprise/technical-requirements/#postgresql-secrets"

		// docsTLSSecrets is the URL to the documentation for TLS secrets.
		//
		// nolint:gosec
		docsTLSSecrets = "https://developer.alpha-sense.com/enterprise/technical-requirements/#tls-secrets"

		// docsSMTPSecrets is the URL to the documentation for SMTP secrets.
		//
		// nolint:gosec
		docsSMTPSecrets = "https://developer.alpha-sense.com/enterprise/technical-requirements/#smtp-credentials-for-email-sending"

		// docsSSOSecrets is the URL to the documentation for SSO secrets.
		//
		// nolint:gosec
		docsSSOSecrets = "https://developer.alpha-sense.com/enterprise/technical-requirements/#sso-secret"

		// docsAWSOIDC is the URL to the documentation for AWS OIDC.
		docsAWSOIDC = "https://developer.alpha-sense.com/enterprise/technical-requirements/aws#oidc-provider-for-iam-role-for-service-account"

		// docsAzureCrossplaneMI is the URL to the documentation for Azure Crossplane Managed Identity.
		docsAzureCrossplaneMI = "https://developer.alpha-sense.com/enterprise/technical-requirements/azure#crossplane-managed-identity"
	)

	c.logger.SetFormatter(log.JSONFormatter)

	c.logger.Debugf(logMsgPodStarted, constant.AppName)

	envConfigBase64 := os.Getenv(envVarEnvConfig)
	if envConfigBase64 == constant.EmptyString {
		c.logger.Fatal(pkgerrors.NewEnvVarIsNotSetOrEmpty(envVarEnvConfig))
	}

	envConfigBytes, err := base64.StdEncoding.DecodeString(envConfigBase64)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToDecodeEnvConfig, err))
	}

	envConfig, err := envconfig.NewFromBytes(envConfigBytes)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToReadEnvConfig, err))
	}

	c.logger.Debug(logMsgEnvConfigDecoded)

	googleCloudSDKDockerRepo := os.Getenv(envVarGoogleCloudSDKDockerRepo)
	if googleCloudSDKDockerRepo == constant.EmptyString {
		c.logger.Fatal(pkgerrors.NewEnvVarIsNotSetOrEmpty(envVarGoogleCloudSDKDockerRepo))
	}

	googleCloudSDKDockerImage := os.Getenv(envVarGoogleCloudSDKDockerImage)
	if googleCloudSDKDockerImage == constant.EmptyString {
		c.logger.Fatal(pkgerrors.NewEnvVarIsNotSetOrEmpty(envVarGoogleCloudSDKDockerImage))
	}

	kubeConfig, path, err := kubeutil.Config(constant.EmptyString)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToGetKubeConfig, err))
	}

	c.logger.Debugf(logMsgKubeLoadedConfig, path)

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToCreateKubernetesClientset, err))
	}

	c.logger.Debug(logMsgKubeClientsetCreated)

	vcloud := cloud.Cloud(envConfig.Spec.CloudSpec.Provider)

	ctx := context.Background()

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
			gcpcloudutil.ServiceAccountAnnotationKey: gcpcloudutil.ServiceAccountAnnotation(envConfig.Spec.ClusterName, envConfig.Spec.CloudSpec.GCP.ProjectID),
		}
	}

	if _, err = clientset.CoreV1().ServiceAccounts(constant.NamespaceCrossplane).Create(
		ctx, sa, metav1.CreateOptions{},
	); err != nil && !k8serrors.IsAlreadyExists(err) {
		c.logger.Fatal(multierr.Combine(errFailedToEnsureServiceAccount, err))
	}

	c.logger.Debugf(logMsgServiceAccountEnsured, constant.NamespaceCrossplane, serviceAccountName)

	httpClient := http.DefaultClient

	checker := cloudchecker.New(c.logger, vcloud, envConfig, clientset, httpClient)

	var jwksURI *string

	rawJWKSURI, err := checker.Handle(ctx)
	if err != nil { // nolint:nestif
		// We don't use c.logger.Fatal() as it will exit the program immediately, and we want to output additional information after logging the fatal error.
		c.logger.Log(log.FatalLevel, multierr.Combine(errFailedToCheckInfrastructure, err))

		docMap := map[error][]string{
			cloudchecker.ErrFailedToCheckStorageClass: {docsPersistentVolumes},
			cloudchecker.ErrFailedToCheckMySQL:        {docsMySQLDatabaseCluster, docsMySQLSecrets},
			cloudchecker.ErrFailedToCheckPostgreSQL:   {docsPostgreSQLDatabaseCluster, docsPostgreSQLSecrets},
			cloudchecker.ErrFailedToCheckTLS:          {docsTLSSecrets},
			cloudchecker.ErrFailedToCheckSMTP:         {docsSMTPSecrets},
			cloudchecker.ErrFailedToCheckSSO:          {docsSSOSecrets},
			cloudchecker.ErrFailedToCheckOIDCURL:      {}, // Special case, docs per cloud provider.
		}

		var targetErr error

		for k := range docMap {
			if errors.Is(err, k) {
				targetErr = k

				break
			}
		}

		if targetErr == nil {
			c.logger.Fatal(errUnknownError)
		}

		if docs, exists := docMap[targetErr]; exists {
			if errors.Is(err, cloudchecker.ErrFailedToCheckOIDCURL) {
				if vcloud == cloud.AWS {
					c.logRelatedDocumentation(docsAWSOIDC)
				} else if vcloud == cloud.Azure {
					c.logRelatedDocumentation(docsAzureCrossplaneMI)
				}
			} else {
				c.logRelatedDocumentation(docs...)
			}
		}

		os.Exit(1)
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
		concreteCloudChecker = gcpchecker.New(c.logger, envConfig, clientset, googleCloudSDKDockerRepo, googleCloudSDKDockerImage)
	}

	if _, err := concreteCloudChecker.Handle(ctx); err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToCheckInfrastructure, err))
	}

	c.logger.Info(logMsgInfraCheckCompletedSuccessfully)
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
		Long: `Pod checks the infrastructure of the cluster where it is running on.

When running this command, provide the environment configuration as a base64 encoded YAML file via the ENVCONFIG environment variable.

You are not supposed to run this command manually, unless you know what you are doing.`,
		Run:    cmd.run,
		Hidden: true,
	}
}
