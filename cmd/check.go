// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/k8s/kubeutil"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var (
	// errFailedToCreateServiceAccount is the error that is returned when the service account cannot be created.
	errFailedToCreateServiceAccount = errors.New("failed to create ServiceAccount")

	// errFailedToCreateRole is the error that is returned when the role cannot be created.
	errFailedToCreateRole = errors.New("failed to create Role")

	// errFailedToCreateClusterRole is the error that is returned when the cluster role cannot be created.
	errFailedToCreateClusterRole = errors.New("failed to create ClusterRole")

	// errFailedToCreateRoleBinding is the error that is returned when the role binding cannot be created.
	errFailedToCreateRoleBinding = errors.New("failed to create RoleBinding")

	// errFailedToCreateClusterRoleBinding is the error that is returned when the cluster role binding cannot be created.
	errFailedToCreateClusterRoleBinding = errors.New("failed to create ClusterRoleBinding")

	// errFailedToMarshalEnvConfig is the error that is returned when the environment configuration cannot be marshaled.
	errFailedToMarshalEnvConfig = errors.New("failed to marshal environment configuration")

	// errFailedToCreatePod is the error that is returned when the pod cannot be created.
	errFailedToCreatePod = errors.New("failed to create Pod")

	// errFailedToDeletePod is the error that is returned when the pod cannot be deleted.
	errFailedToDeletePod = errors.New("failed to delete Pod")

	// errFailedToDeleteRoleBinding is the error that is returned when the role binding cannot be deleted.
	errFailedToDeleteRoleBinding = errors.New("failed to delete RoleBinding")

	// errFailedToDeleteRole is the error that is returned when the role cannot be deleted.
	errFailedToDeleteRole = errors.New("failed to delete Role")

	// errFailedToDeleteServiceAccount is the error that is returned when the service account cannot be deleted.
	errFailedToDeleteServiceAccount = errors.New("failed to delete ServiceAccount")
)

const (
	// flagKubeConfig is the name of the flag for the Kubernetes configuration file.
	flagKubeConfig = "kubeconfig"

	// flagCleanupOnly is the name of the flag for the cleanup only flag.
	flagCleanupOnly = "cleanup-only"

	// flagDockerRepo is the name of the flag for the Docker repository.
	flagDockerRepo = "docker-repo"
	// flagDockerImage is the name of the flag for the Docker image.
	flagDockerImage = "docker-image"
	// flagImagePullSecret is the name of the flag for the image pull secret.
	flagImagePullSecret = "image-pull-secret" // nolint:gosec
)

// namespaceDefault is the default namespace.
const namespaceDefault = "default"

// constRoleNamespaces is the list of namespaces for the roles.
//
// Do not modify this variable, it is supposed to be constant.
var constRoleNamespaces = []string{
	constant.NamespaceAlphaSense,
	constant.NamespaceCrossplane,
	constant.NamespaceMySQL,
	constant.NamespacePlatform,
}

// checkCmd is the command to check the infrastructure.
type checkCmd struct {
	// logger is the logger.
	logger *log.Logger
	// cobraCmd is the Cobra command.
	cobraCmd *cobra.Command

	// envConfig is the environment configuration.
	envConfig *envconfig.EnvConfig
	// kubeConfig is the Kubernetes configuration.
	kubeConfig *rest.Config

	// clientset is the Kubernetes clientset.
	clientset *kubernetes.Clientset
	// clientsetSA is the Kubernetes clientset for the ServiceAccount.
	clientsetSA typedcorev1.ServiceAccountInterface
	// clientsetNamespace is the Kubernetes clientset for the Namespace.
	clientsetNamespace typedcorev1.NamespaceInterface
	// clientsetPod is the Kubernetes clientset for the Pod.
	clientsetPod typedcorev1.PodInterface
}

var _ cmd = &checkCmd{}

// setupClientsets sets up the clientsets.
func (c *checkCmd) setupClientsets() (err error) {
	c.clientset, err = kubernetes.NewForConfig(c.kubeConfig)
	if err != nil {
		return multierr.Combine(errFailedToCreateKubernetesClientset, err)
	}

	c.clientsetSA = c.clientset.CoreV1().ServiceAccounts(namespaceDefault)

	c.clientsetNamespace = c.clientset.CoreV1().Namespaces()

	c.clientsetPod = c.clientset.CoreV1().Pods(namespaceDefault)

	return
}

// createServiceAccount creates the service account.
func (c *checkCmd) createServiceAccount(ctx context.Context, serviceAccountName string) error {
	// logMsgServiceAccountCreated is the message that is logged when the service account is created.
	const logMsgServiceAccountCreated = "created %s/%s ServiceAccount"

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespaceDefault,
		},
	}

	if _, err := c.clientsetSA.Create(ctx, serviceAccount, metav1.CreateOptions{}); err != nil {
		return multierr.Combine(errFailedToCreateServiceAccount, err)
	}

	c.logger.Logf(log.InfoLevel, logMsgServiceAccountCreated, namespaceDefault, serviceAccount.Name)

	return nil
}

// createRoles creates the roles.
//
// nolint:funlen
func (c *checkCmd) createRoles(ctx context.Context, roleName string) error {
	const (
		// logMsgRoleCreated is the message that is logged when the role is created.
		logMsgRoleCreated = "created %s/%s Role"

		// logMsgClusterRoleCreated is the message that is logged when the cluster role is created.
		logMsgClusterRoleCreated = "created %s ClusterRole"
	)

	namespacePolicyRules := []struct {
		namespace string
		rules     []rbacv1.PolicyRule
	}{
		{constant.NamespaceAlphaSense, []rbacv1.PolicyRule{
			{APIGroups: []string{constant.EmptyString}, Resources: []string{"secrets"}, Verbs: []string{rbacv1.VerbAll}},
		}},
		{constant.NamespaceCrossplane, []rbacv1.PolicyRule{
			{APIGroups: []string{constant.EmptyString}, Resources: []string{"pods"}, Verbs: []string{rbacv1.VerbAll}},
			{APIGroups: []string{constant.EmptyString}, Resources: []string{"pods/log"}, Verbs: []string{rbacv1.VerbAll}},
			{APIGroups: []string{constant.EmptyString}, Resources: []string{"pods"}, Verbs: []string{rbacv1.VerbAll}},
			{APIGroups: []string{constant.EmptyString}, Resources: []string{"pods/log"}, Verbs: []string{rbacv1.VerbAll}},
			{APIGroups: []string{constant.EmptyString}, Resources: []string{"serviceaccounts"}, Verbs: []string{rbacv1.VerbAll}},
			{APIGroups: []string{constant.EmptyString}, Resources: []string{"serviceaccounts/token"}, Verbs: []string{rbacv1.VerbAll}},
		}},
		{constant.NamespaceMySQL, []rbacv1.PolicyRule{
			{APIGroups: []string{constant.EmptyString}, Resources: []string{"secrets"}, Verbs: []string{rbacv1.VerbAll}}},
		},
		{constant.NamespacePlatform, []rbacv1.PolicyRule{
			{APIGroups: []string{constant.EmptyString}, Resources: []string{"secrets"}, Verbs: []string{rbacv1.VerbAll}}},
		},
	}

	clusterPolicyRules := []rbacv1.PolicyRule{
		{APIGroups: []string{constant.EmptyString}, Resources: []string{"namespaces"}, Verbs: []string{rbacv1.VerbAll}},
		{APIGroups: []string{constant.EmptyString}, Resources: []string{"nodes"}, Verbs: []string{rbacv1.VerbAll}},
	}

	for _, pair := range namespacePolicyRules {
		role := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      roleName,
				Namespace: pair.namespace,
			},
			Rules: pair.rules,
		}

		if _, err := c.clientset.RbacV1().Roles(pair.namespace).Create(ctx, role, metav1.CreateOptions{}); err != nil {
			return multierr.Combine(errFailedToCreateRole, err)
		}

		c.logger.Logf(log.InfoLevel, logMsgRoleCreated, pair.namespace, role.Name)
	}

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Rules: clusterPolicyRules,
	}

	if _, err := c.clientset.RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{}); err != nil {
		return multierr.Combine(errFailedToCreateClusterRole, err)
	}

	c.logger.Logf(log.InfoLevel, logMsgClusterRoleCreated, clusterRole.Name)

	return nil
}

// createRoleBindings creates the role bindings.
func (c *checkCmd) createRoleBindings(ctx context.Context, serviceAccountName string, roleBindingName string, roleName string) error {
	const (
		// logMsgRoleBindingCreated is the message that is logged when the role binding is created.
		logMsgRoleBindingCreated = "created %s/%s RoleBinding"

		// logMsgClusterRoleBindingCreated is the message that is logged when the cluster role binding is created.
		logMsgClusterRoleBindingCreated = "created %s ClusterRoleBinding"
	)

	// constSubjects is the subjects for the role bindings.
	//
	// Do not modify this variable, it is supposed to be constant.
	constSubjects := []rbacv1.Subject{{
		Kind:      rbacv1.ServiceAccountKind,
		Name:      serviceAccountName,
		Namespace: namespaceDefault,
	}}

	for _, ns := range constRoleNamespaces {
		if _, err := c.clientset.RbacV1().RoleBindings(ns).Create(ctx, &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      roleBindingName,
				Namespace: ns,
			},
			Subjects: constSubjects,
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "Role",
				Name:     roleName,
			},
		}, metav1.CreateOptions{}); err != nil {
			return multierr.Combine(errFailedToCreateRoleBinding, err)
		}

		c.logger.Logf(log.InfoLevel, logMsgRoleBindingCreated, ns, roleBindingName)
	}

	if _, err := c.clientset.RbacV1().ClusterRoleBindings().Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleBindingName,
		},
		Subjects: constSubjects,
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     roleName,
		},
	}, metav1.CreateOptions{}); err != nil {
		return multierr.Combine(errFailedToCreateClusterRoleBinding, err)
	}

	c.logger.Logf(log.InfoLevel, logMsgClusterRoleBindingCreated, roleBindingName)

	return nil
}

// createPod creates the pod.
func (c *checkCmd) createPod(ctx context.Context, serviceAccountName string) error {
	envConfigBytes, err := yaml.Marshal(c.envConfig)
	if err != nil {
		return multierr.Combine(errFailedToMarshalEnvConfig, err)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.AppName,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccountName,
			Containers: []corev1.Container{{
				Name: constant.AppName,
				Image: strings.Join(
					[]string{
						util.Flag(c.cobraCmd, flagDockerRepo),
						util.Flag(c.cobraCmd, flagDockerImage),
					},
					string(constant.HTTPPathSeparator),
				),
				Env: []corev1.EnvVar{{
					Name:  envVarEnvConfig,
					Value: base64.StdEncoding.EncodeToString(envConfigBytes),
				}},
				ImagePullPolicy: corev1.PullAlways,
			}},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	imagePullSecretName := util.Flag(c.cobraCmd, flagImagePullSecret)

	if imagePullSecretName != constant.EmptyString {
		pod.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{
			Name: imagePullSecretName,
		}}
	}

	if _, err = c.clientsetPod.Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return multierr.Combine(errFailedToCreatePod, err)
	}

	c.logger.Logf(log.InfoLevel, constant.LogMsgPodCreated, namespaceDefault, constant.AppName)

	return nil
}

// printPodLogs prints the pod logs.
func (c *checkCmd) printPodLogs(ctx context.Context) error {
	logs, err := kubeutil.PodLogs(ctx, c.logger, c.clientset, namespaceDefault, constant.AppName)
	if err != nil {
		return err
	}

	// logEntry is the struct that represents a log entry.
	type logEntry struct {
		// Timestamp is the Timestamp of the log entry.
		Timestamp string `json:"time"`
		// Level is the Level of the log entry.
		Level string `json:"level"`
		// Message is the Message of the log entry.
		Message string `json:"msg"`
	}

	for _, logStr := range logs {
		var e logEntry

		if err := json.Unmarshal([]byte(logStr), &e); err != nil {
			return err
		}

		if e.Level == constant.EmptyString || e.Message == constant.EmptyString || e.Timestamp == constant.EmptyString {
			continue
		}

		level, err := log.ParseLevel(e.Level)
		if err != nil {
			return err
		}

		parsedTime, err := time.Parse(log.DefaultTimeFormat, e.Timestamp)
		if err != nil {
			return err
		}

		c.logger.SetTimeFunction(func(_ time.Time) time.Time { return parsedTime })

		c.logger.Log(level, e.Message)

		if level == log.FatalLevel {
			os.Exit(1)
		}
	}

	// Reset the time function to the default one, converting to UTC.
	c.logger.SetTimeFunction(constant.LogDefaultTimeFunc)

	return nil
}

// cleanupResources cleans up the resources.
//
// nolint:funlen
func (c *checkCmd) cleanupResources(ctx context.Context, roleBindingName string, roleName string, serviceAccountName string, allowNotFound bool) error {
	const (
		// logMsgClusterRoleBindingDeleted is the message that is logged when the cluster role binding is deleted.
		logMsgClusterRoleBindingDeleted = "deleted %s ClusterRoleBinding"

		// logMsgClusterRoleDeleted is the message that is logged when the cluster role is deleted.
		logMsgClusterRoleDeleted = "deleted %s ClusterRole"

		// logMsgRoleBindingDeleted is the message that is logged when the role binding is deleted.
		logMsgRoleBindingDeleted = "deleted %s/%s RoleBinding"

		// logMsgRoleDeleted is the message that is logged when the role is deleted.
		logMsgRoleDeleted = "deleted %s/%s Role"

		// logMsgServiceAccountDeleted is the message that is logged when the service account is deleted.
		logMsgServiceAccountDeleted = "deleted %s/%s ServiceAccount"
	)

	pod, err := c.clientsetPod.Get(ctx, constant.AppName, metav1.GetOptions{})
	if err != nil && (!allowNotFound && !k8serrors.IsNotFound(err)) {
		return multierr.Combine(kubeutil.ErrFailedToGetPod, err)
	}

	if err = c.clientsetPod.Delete(ctx, constant.AppName, metav1.DeleteOptions{}); err != nil && !allowNotFound && !k8serrors.IsNotFound(err) {
		return multierr.Combine(errFailedToDeletePod, err)
	}

	c.logger.Logf(log.InfoLevel, constant.LogMsgPodDeleted, namespaceDefault, constant.AppName)

	if err = c.clientset.RbacV1().ClusterRoleBindings().Delete(
		ctx,
		roleBindingName,
		metav1.DeleteOptions{},
	); err != nil && !allowNotFound && !k8serrors.IsNotFound(err) {
		return multierr.Combine(errFailedToDeleteRoleBinding, err)
	}

	c.logger.Logf(log.InfoLevel, logMsgClusterRoleBindingDeleted, roleBindingName)

	if err = c.clientset.RbacV1().ClusterRoles().Delete(
		ctx,
		roleName,
		metav1.DeleteOptions{},
	); err != nil && !allowNotFound && !k8serrors.IsNotFound(err) {
		return multierr.Combine(errFailedToDeleteRole, err)
	}

	c.logger.Logf(log.InfoLevel, logMsgClusterRoleDeleted, roleName)

	for _, ns := range constRoleNamespaces {
		if err = c.clientset.RbacV1().RoleBindings(ns).Delete(
			ctx,
			roleBindingName,
			metav1.DeleteOptions{},
		); err != nil && !allowNotFound && !k8serrors.IsNotFound(err) {
			return multierr.Combine(errFailedToDeleteRoleBinding, err)
		}

		c.logger.Logf(log.InfoLevel, logMsgRoleBindingDeleted, ns, roleBindingName)

		if err = c.clientset.RbacV1().Roles(ns).Delete(
			ctx,
			roleName,
			metav1.DeleteOptions{},
		); err != nil && !allowNotFound && !k8serrors.IsNotFound(err) {
			return multierr.Combine(errFailedToDeleteRole, err)
		}

		c.logger.Logf(log.InfoLevel, logMsgRoleDeleted, ns, roleName)
	}

	if err = c.clientsetSA.Delete(ctx, serviceAccountName, metav1.DeleteOptions{}); err != nil && !allowNotFound && !k8serrors.IsNotFound(err) {
		return multierr.Combine(errFailedToDeleteServiceAccount, err)
	}

	c.logger.Logf(log.InfoLevel, logMsgServiceAccountDeleted, namespaceDefault, serviceAccountName)

	if pod != nil && !allowNotFound && pod.Status.Phase == corev1.PodFailed {
		os.Exit(1)
	}

	return nil
}

// Run is the run function for the Check command.
//
// nolint:funlen
func (c *checkCmd) Run(_ *cobra.Command, args []string) {
	const (
		// logMsgInfraCheckStarted is the message that is logged when the infrastructure check starts.
		logMsgInfraCheckStarted = "started infrastructure check"

		// logMsgEnvConfigRead is the message that is logged when the environment configuration is read from the specified path.
		logMsgEnvConfigRead = "read environment configuration from %s"
	)

	c.logger.Log(log.InfoLevel, logMsgInfraCheckStarted)

	firstStepFile := args[0]

	c.logger.Logf(log.InfoLevel, logMsgEnvConfigRead, firstStepFile)

	var err error

	c.envConfig, err = envconfig.NewFromPath(firstStepFile)
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToReadEnvConfig, err))
	}

	var path string

	c.kubeConfig, path, err = kubeutil.Config(util.Flag(c.cobraCmd, flagKubeConfig))
	if err != nil {
		c.logger.Fatal(multierr.Combine(errFailedToGetKubeConfig, err))
	}

	c.logger.Logf(log.InfoLevel, logMsgKubeLoadedConfig, path)

	serviceAccountName := fmt.Sprintf("%s-sa", constant.AppName)

	roleName := fmt.Sprintf("%s-role", constant.AppName)

	roleBindingName := fmt.Sprintf("%s-rolebinding", constant.AppName)

	ctx := context.Background()

	if err = c.setupClientsets(); err != nil {
		c.logger.Fatal(err)
	}

	if util.FlagBool(c.cobraCmd, flagCleanupOnly) {
		if err = c.cleanupResources(ctx, roleBindingName, roleName, serviceAccountName, true); err != nil {
			c.logger.Fatal(err)
		}

		return
	}

	c.logger.Log(log.InfoLevel, logMsgKubeClientsetCreated)

	if err = c.createServiceAccount(ctx, serviceAccountName); err != nil {
		c.logger.Fatal(err)
	}

	if err = c.createRoles(ctx, roleName); err != nil {
		c.logger.Fatal(err)
	}

	if err = c.createRoleBindings(ctx, serviceAccountName, roleBindingName, roleName); err != nil {
		c.logger.Fatal(err)
	}

	if err = c.createPod(ctx, serviceAccountName); err != nil {
		c.logger.Fatal(err)
	}

	_, err = kubeutil.WaitForPodToSucceedOrFail(ctx, c.logger, c.clientset, namespaceDefault, constant.AppName)
	if err != nil {
		c.logger.Fatal(err)
	}

	if err = c.printPodLogs(ctx); err != nil {
		c.logger.Fatal(err)
	}

	if err = c.cleanupResources(ctx, roleBindingName, roleName, serviceAccountName, false); err != nil {
		c.logger.Fatal(err)
	}
}

// newCheckCmd returns a new checkCmd.
func newCheckCmd(logger *log.Logger, cobraCmd *cobra.Command) *checkCmd {
	return &checkCmd{
		logger:   logger,
		cobraCmd: cobraCmd,
	}
}

// Check returns a Cobra command to check the infrastructure.
func Check(logger *log.Logger) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "check <first_step_file>",
		Short: "Check the infrastructure",
		Long: fmt.Sprintf(
			`Check reviews the infrastructure in your cloud environment to ensure it is ready for deployment.

You may specify the Kubernetes configuration file to use by setting the --%s flag or by setting the KUBECONFIG environment variable.
If you do not specify the Kubernetes configuration file, the command will use the default Kubernetes configuration file located at your home directory.`,
			flagKubeConfig,
		),
		Args: cobra.ExactArgs(1),
	}

	cmd := newCheckCmd(logger, cobraCmd)

	cobraCmd.Run = cmd.Run

	const (
		// defaultDockerRepo is the default repository to use for the pod image.
		defaultDockerRepo = ""
	)

	var (
		// defaultDockerImage is the default image to use for the pod.
		defaultDockerImage = fmt.Sprintf("%s-pod:latest", constant.AppName)
	)

	cobraCmd.Flags().String(
		flagKubeConfig,
		constant.EmptyString,
		"path to the Kubernetes configuration file to use for the check (or KUBECONFIG environment variable)",
	)

	cobraCmd.Flags().Bool(flagCleanupOnly, false, "only clean up the resources and exit")

	cobraCmd.Flags().String(flagDockerRepo, defaultDockerRepo, "the Docker repository to use for the Pod image")
	cobraCmd.Flags().String(flagDockerImage, defaultDockerImage, "the Docker image to use for the Pod")
	cobraCmd.Flags().String(flagImagePullSecret, constant.EmptyString, "the name of the image pull secret to use for the Pod")

	return cobraCmd
}
