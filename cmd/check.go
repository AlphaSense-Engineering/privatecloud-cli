// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/k8s/kubeutil"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var (
	// errFailedToCreateServiceAccount is the error that is returned when the service account cannot be created.
	errFailedToCreateServiceAccount = errors.New("failed to create ServiceAccount")

	// errFailedToEnsureNamespace is the error that is returned when the namespace cannot be ensured.
	errFailedToEnsureNamespace = errors.New("failed to ensure Namespace")

	// errFailedToCreateRole is the error that is returned when the role cannot be created.
	errFailedToCreateRole = errors.New("failed to create Role")

	// errFailedToCreateRoleBinding is the error that is returned when the role binding cannot be created.
	errFailedToCreateRoleBinding = errors.New("failed to create RoleBinding")

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

// namespaceKubeSystem is the namespace for the Kubernetes system.
const namespaceKubeSystem = "kube-system"

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

	c.clientsetSA = c.clientset.CoreV1().ServiceAccounts(namespaceKubeSystem)

	c.clientsetNamespace = c.clientset.CoreV1().Namespaces()

	c.clientsetPod = c.clientset.CoreV1().Pods(namespaceKubeSystem)

	return
}

// createServiceAccount creates the service account.
func (c *checkCmd) createServiceAccount(ctx context.Context, serviceAccountName string) error {
	// logMsgServiceAccountCreated is the message that is logged when the service account is created.
	const logMsgServiceAccountCreated = "created %s/%s ServiceAccount"

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespaceKubeSystem,
		},
	}

	if _, err := c.clientsetSA.Create(ctx, serviceAccount, metav1.CreateOptions{}); err != nil {
		return multierr.Combine(errFailedToCreateServiceAccount, err)
	}

	log.Printf(logMsgServiceAccountCreated, namespaceKubeSystem, serviceAccount.Name)

	return nil
}

// ensureNamespace ensures the namespace.
func (c *checkCmd) ensureNamespace(ctx context.Context) error {
	// logMsgNamespaceEnsured is the message that is logged when the namespace is ensured.
	const logMsgNamespaceEnsured = "ensured %s Namespace"

	if _, err := c.clientsetNamespace.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.NamespaceCrossplane,
		},
	}, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		return multierr.Combine(errFailedToEnsureNamespace, err)
	}

	log.Printf(logMsgNamespaceEnsured, constant.NamespaceCrossplane)

	return nil
}

// createRoles creates the roles.
func (c *checkCmd) createRoles(ctx context.Context, roleName string) error {
	// logMsgRoleCreated is the message that is logged when the role is created.
	const logMsgRoleCreated = "created %s/%s Role"

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

		log.Printf(logMsgRoleCreated, pair.namespace, role.Name)
	}

	return nil
}

// createRoleBindings creates the role bindings.
func (c *checkCmd) createRoleBindings(ctx context.Context, serviceAccountName string, roleBindingName string, roleName string) error {
	// logMsgRoleBindingCreated is the message that is logged when the role binding is created.
	const logMsgRoleBindingCreated = "created %s/%s RoleBinding"

	for _, ns := range constRoleNamespaces {
		if _, err := c.clientset.RbacV1().RoleBindings(ns).Create(ctx, &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      roleBindingName,
				Namespace: ns,
			},
			Subjects: []rbacv1.Subject{{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      serviceAccountName,
				Namespace: namespaceKubeSystem,
			}},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "Role",
				Name:     roleName,
			},
		}, metav1.CreateOptions{}); err != nil {
			return multierr.Combine(errFailedToCreateRoleBinding, err)
		}

		log.Printf(logMsgRoleBindingCreated, ns, roleBindingName)
	}

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

	log.Printf(constant.LogMsgPodCreated, namespaceKubeSystem, constant.AppName)

	return nil
}

// printPodLogs prints the pod logs.
func (c *checkCmd) printPodLogs(ctx context.Context) error {
	logs, err := kubeutil.PodLogs(ctx, c.clientset, namespaceKubeSystem, constant.AppName)
	if err != nil {
		return err
	}

	log.SetFlags(0)

	log.Println(strings.Join(logs, "\n"))

	log.SetFlags(constant.LogFlags)

	return nil
}

// cleanupResources cleans up the resources.
//
// nolint:funlen
func (c *checkCmd) cleanupResources(ctx context.Context, roleBindingName string, roleName string, serviceAccountName string, allowNotFound bool) error {
	const (
		// logMsgRoleBindingDeleted is the message that is logged when the role binding is deleted.
		logMsgRoleBindingDeleted = "deleted %s/%s RoleBinding"

		// logMsgRoleDeleted is the message that is logged when the role is deleted.
		logMsgRoleDeleted = "deleted %s/%s Role"

		// logMsgServiceAccountDeleted is the message that is logged when the service account is deleted.
		logMsgServiceAccountDeleted = "deleted %s/%s ServiceAccount"
	)

	pod, err := c.clientsetPod.Get(ctx, constant.AppName, metav1.GetOptions{})
	if err != nil && (!allowNotFound && !apierrors.IsNotFound(err)) {
		return multierr.Combine(kubeutil.ErrFailedToGetPod, err)
	}

	if err = c.clientsetPod.Delete(ctx, constant.AppName, metav1.DeleteOptions{}); err != nil && !allowNotFound && !apierrors.IsNotFound(err) {
		return multierr.Combine(errFailedToDeletePod, err)
	}

	log.Printf(constant.LogMsgPodDeleted, namespaceKubeSystem, constant.AppName)

	for _, ns := range constRoleNamespaces {
		if err = c.clientset.RbacV1().RoleBindings(ns).Delete(
			ctx,
			roleBindingName,
			metav1.DeleteOptions{},
		); err != nil && !allowNotFound && !apierrors.IsNotFound(err) {
			return multierr.Combine(errFailedToDeleteRoleBinding, err)
		}

		log.Printf(logMsgRoleBindingDeleted, ns, roleBindingName)

		if err = c.clientset.RbacV1().Roles(ns).Delete(
			ctx,
			roleName,
			metav1.DeleteOptions{},
		); err != nil && !allowNotFound && !apierrors.IsNotFound(err) {
			return multierr.Combine(errFailedToDeleteRole, err)
		}

		log.Printf(logMsgRoleDeleted, ns, roleName)
	}

	if err = c.clientsetSA.Delete(ctx, serviceAccountName, metav1.DeleteOptions{}); err != nil && !allowNotFound && !apierrors.IsNotFound(err) {
		return multierr.Combine(errFailedToDeleteServiceAccount, err)
	}

	log.Printf(logMsgServiceAccountDeleted, namespaceKubeSystem, serviceAccountName)

	if pod != nil && !allowNotFound && pod.Status.Phase == corev1.PodFailed {
		os.Exit(1)
	}

	return nil
}

// Run is the run function for the Check command.
//
// nolint:funlen
func (c *checkCmd) Run(cobraCmd *cobra.Command, args []string) {
	const (
		// logMsgInfraCheckStarted is the message that is logged when the infrastructure check starts.
		logMsgInfraCheckStarted = "started infrastructure check"

		// logMsgEnvConfigRead is the message that is logged when the environment configuration is read from the specified path.
		logMsgEnvConfigRead = "read environment configuration from %s"
	)

	c.cobraCmd = cobraCmd

	log.SetFlags(constant.LogFlags)

	log.Println(logMsgInfraCheckStarted)

	firstStepFile := args[0]

	log.Printf(logMsgEnvConfigRead, firstStepFile)

	var err error

	c.envConfig, err = envconfig.NewFromPath(firstStepFile)
	if err != nil {
		log.Fatalln(multierr.Combine(errFailedToReadEnvConfig, err))

		return
	}

	var path string

	c.kubeConfig, path, err = kubeutil.Config(util.Flag(c.cobraCmd, flagKubeConfig))
	if err != nil {
		log.Fatalln(multierr.Combine(errFailedToGetKubeConfig, err))
	}

	log.Printf(logMsgKubeLoadedConfig, path)

	serviceAccountName := fmt.Sprintf("%s-sa", constant.AppName)

	roleName := fmt.Sprintf("%s-role", constant.AppName)

	roleBindingName := fmt.Sprintf("%s-rolebinding", constant.AppName)

	ctx := context.Background()

	if err = c.setupClientsets(); err != nil {
		log.Fatalln(err)
	}

	if util.FlagBool(c.cobraCmd, flagCleanupOnly) {
		if err = c.cleanupResources(ctx, roleBindingName, roleName, serviceAccountName, true); err != nil {
			log.Fatalln(err)
		}

		return
	}

	log.Println(logMsgKubeClientsetCreated)

	if err = c.createServiceAccount(ctx, serviceAccountName); err != nil {
		log.Fatalln(err)
	}

	if err = c.ensureNamespace(ctx); err != nil {
		log.Fatalln(err)
	}

	if err = c.createRoles(ctx, roleName); err != nil {
		log.Fatalln(err)
	}

	if err = c.createRoleBindings(ctx, serviceAccountName, roleBindingName, roleName); err != nil {
		log.Fatalln(err)
	}

	if err = c.createPod(ctx, serviceAccountName); err != nil {
		log.Fatalln(err)
	}

	_, err = kubeutil.WaitForPodToSucceedOrFail(ctx, c.clientset, namespaceKubeSystem, constant.AppName)
	if err != nil {
		log.Fatalln(err)
	}

	if err = c.printPodLogs(ctx); err != nil {
		log.Fatalln(err)
	}

	if err = c.cleanupResources(ctx, roleBindingName, roleName, serviceAccountName, false); err != nil {
		log.Fatalln(err)
	}
}

// newCheckCmd returns a new Check command.
func newCheckCmd() *checkCmd {
	return &checkCmd{}
}

// Check returns a Cobra command to check the infrastructure.
func Check() *cobra.Command {
	cmd := newCheckCmd()

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
		Run:  cmd.Run,
	}

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
