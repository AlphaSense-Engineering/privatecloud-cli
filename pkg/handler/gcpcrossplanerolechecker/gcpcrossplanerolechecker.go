// Package gcpcrossplanerolechecker is the package that contains the check functions for GCP Crossplane role.
package gcpcrossplanerolechecker

import (
	"context"
	"errors"
	"strings"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	selferrors "github.com/AlphaSense-Engineering/privatecloud-installer/pkg/errors"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/k8s/kubeutil"
	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// errMoreThanOneLogLine is the error that is returned when we expect 1 log line but got more than 1.
var errMoreThanOneLogLine = errors.New("got more than 1 log line")

const (
	// podName is the name of the pod that checks the GCP Crossplane role.
	podName = "gcp-crossplane-role-checker"

	// bashScript is the bash script that checks the GCP Crossplane role.
	bashScript = `EMAIL=$(gcloud auth list --filter=status:ACTIVE --format="value(account)")
PROJECT_ID=$(gcloud config get-value project)

ROLES=$(gcloud projects get-iam-policy "$PROJECT_ID" \
  --flatten="bindings[].members" --filter="bindings.members:$EMAIL" --format="value(bindings.role)")

ROLE_COUNT=0
SELECTED_ROLE_ID=""

for ROLE in $ROLES; do
  ROLE_ID=$(echo "$ROLE" | sed 's|.*/||')
  if [[ "$ROLE_ID" = uxp_provider* ]]; then
    ROLE_COUNT=$((ROLE_COUNT + 1))
    SELECTED_ROLE_ID="$ROLE_ID"

    if [[ $ROLE_COUNT -gt 1 ]]; then
      echo "More than one uxp_provider role found" >&2
      exit 1
    fi
  fi
done

if [[ $ROLE_COUNT -eq 1 ]]; then
  gcloud iam roles describe "$SELECTED_ROLE_ID" --project="$PROJECT_ID" --format="value(includedPermissions)" || exit 1
  exit 0
fi

echo "No uxp_provider role found" >&2
exit 1`
)

// constExpectedRolePermissions are the expected permissions for the Crossplane role in GCP.
//
// These are listed at https://developer.alpha-sense.com/enterprise/technical-requirements/gcp.
//
// Do not modify this variable, it is supposed to be constant.
var constExpectedRolePermissions = map[string]struct{}{
	"cloudsql.backupRuns.create":            {},
	"cloudsql.backupRuns.delete":            {},
	"cloudsql.backupRuns.get":               {},
	"cloudsql.backupRuns.list":              {},
	"cloudsql.instances.addServerCa":        {},
	"cloudsql.instances.clone":              {},
	"cloudsql.instances.connect":            {},
	"cloudsql.instances.create":             {},
	"cloudsql.instances.createTagBinding":   {},
	"cloudsql.instances.delete":             {},
	"cloudsql.instances.deleteTagBinding":   {},
	"cloudsql.instances.export":             {},
	"cloudsql.instances.failover":           {},
	"cloudsql.instances.get":                {},
	"cloudsql.instances.import":             {},
	"cloudsql.instances.list":               {},
	"cloudsql.instances.listEffectiveTags":  {},
	"cloudsql.instances.listTagBindings":    {},
	"cloudsql.instances.resetSslConfig":     {},
	"cloudsql.instances.restart":            {},
	"cloudsql.instances.restoreBackup":      {},
	"cloudsql.instances.update":             {},
	"cloudsql.users.create":                 {},
	"cloudsql.users.delete":                 {},
	"cloudsql.users.get":                    {},
	"cloudsql.users.list":                   {},
	"cloudsql.users.update":                 {},
	"iam.roles.create":                      {},
	"iam.roles.delete":                      {},
	"iam.roles.get":                         {},
	"iam.roles.list":                        {},
	"iam.roles.undelete":                    {},
	"iam.roles.update":                      {},
	"iam.serviceAccountKeys.create":         {},
	"iam.serviceAccountKeys.delete":         {},
	"iam.serviceAccountKeys.disable":        {},
	"iam.serviceAccountKeys.enable":         {},
	"iam.serviceAccountKeys.get":            {},
	"iam.serviceAccountKeys.list":           {},
	"iam.serviceAccounts.create":            {},
	"iam.serviceAccounts.delete":            {},
	"iam.serviceAccounts.disable":           {},
	"iam.serviceAccounts.enable":            {},
	"iam.serviceAccounts.get":               {},
	"iam.serviceAccounts.getIamPolicy":      {},
	"iam.serviceAccounts.list":              {},
	"iam.serviceAccounts.setIamPolicy":      {},
	"iam.serviceAccounts.undelete":          {},
	"iam.serviceAccounts.update":            {},
	"pubsub.subscriptions.create":           {},
	"pubsub.subscriptions.delete":           {},
	"pubsub.subscriptions.get":              {},
	"pubsub.subscriptions.getIamPolicy":     {},
	"pubsub.subscriptions.list":             {},
	"pubsub.subscriptions.setIamPolicy":     {},
	"pubsub.subscriptions.update":           {},
	"pubsub.topics.attachSubscription":      {},
	"pubsub.topics.create":                  {},
	"pubsub.topics.delete":                  {},
	"pubsub.topics.detachSubscription":      {},
	"pubsub.topics.get":                     {},
	"pubsub.topics.getIamPolicy":            {},
	"pubsub.topics.list":                    {},
	"pubsub.topics.setIamPolicy":            {},
	"pubsub.topics.update":                  {},
	"pubsub.topics.updateTag":               {},
	"resourcemanager.projects.get":          {},
	"resourcemanager.projects.getIamPolicy": {},
	"resourcemanager.projects.setIamPolicy": {},
	"storage.buckets.create":                {},
	"storage.buckets.createTagBinding":      {},
	"storage.buckets.delete":                {},
	"storage.buckets.deleteTagBinding":      {},
	"storage.buckets.enableObjectRetention": {},
	"storage.buckets.get":                   {},
	"storage.buckets.getIamPolicy":          {},
	"storage.buckets.list":                  {},
	"storage.buckets.listEffectiveTags":     {},
	"storage.buckets.listTagBindings":       {},
	"storage.buckets.setIamPolicy":          {},
	"storage.buckets.update":                {},
	"storage.hmacKeys.create":               {},
	"storage.hmacKeys.delete":               {},
	"storage.hmacKeys.get":                  {},
	"storage.hmacKeys.list":                 {},
	"storage.hmacKeys.update":               {},
}

// GCPCrossplaneRoleChecker is the type that contains the check functions for GCP Crossplane role.
type GCPCrossplaneRoleChecker struct {
	// logger is the logger.
	logger *log.Logger
	// envConfig is the environment configuration.
	envConfig *envconfig.EnvConfig
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &GCPCrossplaneRoleChecker{}

// Handle is the function that handles the GCP Crossplane role check.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
//
// nolint:funlen
func (c *GCPCrossplaneRoleChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: constant.NamespaceCrossplane,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: constant.ServiceAccountNameGCP,
			Containers: []corev1.Container{{
				Name:  podName,
				Image: "google/cloud-sdk:latest",
				Command: []string{
					"/bin/bash",
					"-c",
					bashScript,
				}},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	clientsetPod := c.clientset.CoreV1().Pods(constant.NamespaceCrossplane)

	_, err := clientsetPod.Get(ctx, podName, metav1.GetOptions{})
	if err == nil {
		if err := clientsetPod.Delete(ctx, podName, metav1.DeleteOptions{}); err != nil {
			return nil, err
		}

		c.logger.Logf(log.InfoLevel, constant.LogMsgPodDeleted, constant.NamespaceCrossplane, podName)
	} else if !k8serrors.IsNotFound(err) {
		return nil, err
	}

	if _, err := clientsetPod.Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	c.logger.Logf(log.InfoLevel, constant.LogMsgPodCreated, constant.NamespaceCrossplane, podName)

	phase, err := kubeutil.WaitForPodToSucceedOrFail(ctx, c.logger, c.clientset, constant.NamespaceCrossplane, podName)
	if err != nil {
		return nil, err
	}

	logs, err := kubeutil.PodLogs(ctx, c.logger, c.clientset, constant.NamespaceCrossplane, podName)
	if err != nil {
		return nil, err
	}

	if len(logs) > 1 {
		return nil, errMoreThanOneLogLine
	}

	logLine := logs[0]

	if phase == corev1.PodFailed {
		return nil, errors.New(logLine)
	}

	permissions := strings.Split(logLine, ";")

	missingPermissions := []string{}

	for expectedPermission := range constExpectedRolePermissions {
		found := false

		for _, permission := range permissions {
			if permission == expectedPermission {
				found = true

				break
			}
		}

		if !found {
			missingPermissions = append(missingPermissions, expectedPermission)
		}
	}

	if len(missingPermissions) > 0 {
		return nil, selferrors.NewRoleMissingPermissions(missingPermissions)
	}

	if err := clientsetPod.Delete(ctx, podName, metav1.DeleteOptions{}); err != nil {
		return nil, err
	}

	c.logger.Logf(log.InfoLevel, constant.LogMsgPodDeleted, constant.NamespaceCrossplane, podName)

	return nil, nil
}

// New is the function that creates a new GCPCrossplaneRoleChecker.
func New(logger *log.Logger, envConfig *envconfig.EnvConfig, clientset kubernetes.Interface) *GCPCrossplaneRoleChecker {
	return &GCPCrossplaneRoleChecker{
		logger:    logger,
		envConfig: envConfig,
		clientset: clientset,
	}
}
