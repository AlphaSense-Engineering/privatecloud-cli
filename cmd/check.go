// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
	// errFailedToReadEnvConfig is the error that is returned when the environment configuration cannot be read.
	errFailedToReadEnvConfig = errors.New("failed to read environment configuration")

	// errFailedToGetKubeConfig is the error that is returned when the Kubernetes configuration cannot be retrieved.
	errFailedToGetKubeConfig = errors.New("failed to get Kubernetes configuration")

	// errFailedToCreateKubernetesClientset is the error that is returned when the Kubernetes clientset cannot be created.
	errFailedToCreateKubernetesClientset = errors.New("failed to create Kubernetes clientset")

	// errFailedToGetPod is the error that is returned when the pod cannot be retrieved.
	errFailedToGetPod = errors.New("failed to get Pod")

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

	// errFailedToGetPodLogStream is the error that is returned when the pod log stream cannot be retrieved.
	errFailedToGetPodLogStream = errors.New("failed to get Pod log stream")

	// errFailedToReadPodLogStream is the error that is returned when the pod log stream cannot be read.
	errFailedToReadPodLogStream = errors.New("failed to read Pod log stream")

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

	// flagDockerRepo is the name of the flag for the Docker repository.
	flagDockerRepo = "docker-repo"
)

// namespaceKubeSystem is the namespace for the Kubernetes system.
const namespaceKubeSystem = "kube-system"

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
func (c *checkCmd) createServiceAccount(ctx context.Context) (*corev1.ServiceAccount, error) {
	// logMsgServiceAccountCreated is the message that is logged when the service account is created.
	const logMsgServiceAccountCreated = "created %s/%s ServiceAccount"

	serviceAccountName := fmt.Sprintf("%s-sa", constant.AppName)

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespaceKubeSystem,
		},
	}

	_, err := c.clientsetSA.Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return nil, multierr.Combine(errFailedToCreateServiceAccount, err)
	}

	log.Printf(logMsgServiceAccountCreated, namespaceKubeSystem, serviceAccount.Name)

	return serviceAccount, nil
}

// ensureNamespace ensures the namespace.
func (c *checkCmd) ensureNamespace(ctx context.Context) error {
	// logMsgNamespaceEnsured is the message that is logged when the namespace is ensured.
	const logMsgNamespaceEnsured = "ensured %s Namespace"

	_, err := c.clientsetNamespace.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceCrossplane,
		},
	}, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return multierr.Combine(errFailedToEnsureNamespace, err)
	}

	log.Printf(logMsgNamespaceEnsured, namespaceCrossplane)

	return nil
}

// createRoles creates the roles.
func (c *checkCmd) createRoles(ctx context.Context) (*rbacv1.Role, *rbacv1.Role, error) {
	// logMsgRoleCreated is the message that is logged when the role is created.
	const logMsgRoleCreated = "created %s/%s Role"

	roleName := fmt.Sprintf("%s-role", constant.AppName)

	roleCrossplane := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespaceCrossplane,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{constant.EmptyString},
				Resources: []string{"serviceaccounts"},
				Verbs:     []string{rbacv1.VerbAll},
			},
			{
				APIGroups: []string{constant.EmptyString},
				Resources: []string{"serviceaccounts/token"},
				Verbs:     []string{rbacv1.VerbAll},
			},
		},
	}

	_, err := c.clientset.RbacV1().Roles(namespaceCrossplane).Create(ctx, roleCrossplane, metav1.CreateOptions{})
	if err != nil {
		return nil, nil, multierr.Combine(errFailedToCreateRole, err)
	}

	log.Printf(logMsgRoleCreated, namespaceCrossplane, roleCrossplane.Name)

	roleMySQL := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: constant.NamespaceMySQL,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{constant.EmptyString},
				Resources: []string{"secrets"},
				Verbs:     []string{rbacv1.VerbAll},
			},
		},
	}

	_, err = c.clientset.RbacV1().Roles(constant.NamespaceMySQL).Create(ctx, roleMySQL, metav1.CreateOptions{})
	if err != nil {
		return nil, nil, multierr.Combine(errFailedToCreateRole, err)
	}

	log.Printf(logMsgRoleCreated, constant.NamespaceMySQL, roleMySQL.Name)

	return roleCrossplane, roleMySQL, nil
}

// createRoleBindings creates the role bindings.
func (c *checkCmd) createRoleBindings(ctx context.Context, serviceAccount *corev1.ServiceAccount, roleCrossplane *rbacv1.Role, roleMySQL *rbacv1.Role) error {
	// logMsgRoleBindingCreated is the message that is logged when the role binding is created.
	const logMsgRoleBindingCreated = "created %s/%s RoleBinding"

	roleBindingName := fmt.Sprintf("%s-rolebinding", constant.AppName)

	_, err := c.clientset.RbacV1().RoleBindings(namespaceCrossplane).Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: namespaceCrossplane,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccount.Name,
			Namespace: namespaceKubeSystem,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     roleCrossplane.Name,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return multierr.Combine(errFailedToCreateRoleBinding, err)
	}

	log.Printf(logMsgRoleBindingCreated, namespaceCrossplane, roleBindingName)

	_, err = c.clientset.RbacV1().RoleBindings(constant.NamespaceMySQL).Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: constant.NamespaceMySQL,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccount.Name,
			Namespace: namespaceKubeSystem,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     roleMySQL.Name,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return multierr.Combine(errFailedToCreateRoleBinding, err)
	}

	log.Printf(logMsgRoleBindingCreated, constant.NamespaceMySQL, roleBindingName)

	return nil
}

// createPod creates the pod.
func (c *checkCmd) createPod(ctx context.Context, serviceAccount *corev1.ServiceAccount) error {
	// logMsgPodCreated is the message that is logged when the pod is created.
	const logMsgPodCreated = "created %s/%s Pod"

	envConfigBytes, err := yaml.Marshal(c.envConfig)
	if err != nil {
		return multierr.Combine(errFailedToMarshalEnvConfig, err)
	}

	_, err = c.clientsetPod.Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.AppName,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccount.Name,
			Containers: []corev1.Container{{
				Name: constant.AppName,
				Image: strings.Join([]string{
					util.Flag(c.cobraCmd, flagDockerRepo),
					fmt.Sprintf("%s-pod:0.0.1", constant.AppName),
				}, string(constant.HTTPPathSeparator)),
				Env: []corev1.EnvVar{{
					Name:  envVarEnvConfig,
					Value: base64.StdEncoding.EncodeToString(envConfigBytes),
				}},
				ImagePullPolicy: corev1.PullAlways,
			}},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return multierr.Combine(errFailedToCreatePod, err)
	}

	log.Printf(logMsgPodCreated, namespaceKubeSystem, constant.AppName)

	return nil
}

// waitForPodToRun waits for the pod to run.
func (c *checkCmd) waitForPodToRun(ctx context.Context) error {
	const (
		// logMsgPodWaitingToRun is the message that is logged when the pod is waiting to run.
		logMsgPodWaitingToRun = "waiting for %s/%s Pod to run..."

		// logMsgPodRunning is the message that is logged when the pod is running.
		logMsgPodRunning = "%s/%s Pod is running"
	)

	log.Printf(logMsgPodWaitingToRun, namespaceKubeSystem, constant.AppName)

	for {
		pod, err := c.clientsetPod.Get(ctx, constant.AppName, metav1.GetOptions{})
		if err != nil {
			return multierr.Combine(errFailedToGetPod, err)
		}

		if pod.Status.Phase == corev1.PodRunning {
			log.Printf(logMsgPodRunning, namespaceKubeSystem, constant.AppName)

			break
		}

		time.Sleep(time.Second)
	}

	return nil
}

// streamPodLogs streams the pod logs.
func (c *checkCmd) streamPodLogs(ctx context.Context) error {
	// logMsgPodLogStreamRetrieved is the message that is logged when the pod log stream is retrieved.
	const logMsgPodLogStreamRetrieved = "retrieved log stream for %s/%s Pod, printing..."

	req := c.clientsetPod.GetLogs(constant.AppName, &corev1.PodLogOptions{})

	podLogStream, err := req.Stream(ctx)
	if err != nil {
		return multierr.Combine(errFailedToGetPodLogStream, err)
	}
	defer podLogStream.Close() // nolint:errcheck

	log.Printf(logMsgPodLogStreamRetrieved, namespaceKubeSystem, constant.AppName)

	log.SetFlags(0)

	scanner := bufio.NewScanner(podLogStream)

	for scanner.Scan() {
		trimmedLine := strings.TrimSpace(scanner.Text())

		if trimmedLine == constant.EmptyString {
			continue
		}

		log.Println(trimmedLine)
	}

	if err := scanner.Err(); err != nil {
		return multierr.Combine(errFailedToReadPodLogStream, err)
	}

	log.SetFlags(constant.LogFlags)

	return nil
}

// cleanupResources cleans up the resources.
func (c *checkCmd) cleanupResources(ctx context.Context, roleCrossplane *rbacv1.Role, roleMySQL *rbacv1.Role, serviceAccount *corev1.ServiceAccount) error {
	// logMsgResourcesCleanedUp is the message that is logged when the resources are cleaned up.
	const logMsgResourcesCleanedUp = "resources cleaned up"

	pod, err := c.clientsetPod.Get(ctx, constant.AppName, metav1.GetOptions{})
	if err != nil {
		return multierr.Combine(errFailedToGetPod, err)
	}

	if err = c.clientsetPod.Delete(ctx, constant.AppName, metav1.DeleteOptions{}); err != nil {
		return multierr.Combine(errFailedToDeletePod, err)
	}

	roleBindingName := fmt.Sprintf("%s-rolebinding", constant.AppName)

	if err = c.clientset.RbacV1().RoleBindings(constant.NamespaceMySQL).Delete(ctx, roleBindingName, metav1.DeleteOptions{}); err != nil {
		return multierr.Combine(errFailedToDeleteRoleBinding, err)
	}

	if err = c.clientset.RbacV1().RoleBindings(namespaceCrossplane).Delete(ctx, roleBindingName, metav1.DeleteOptions{}); err != nil {
		return multierr.Combine(errFailedToDeleteRoleBinding, err)
	}

	if err = c.clientset.RbacV1().Roles(constant.NamespaceMySQL).Delete(ctx, roleMySQL.Name, metav1.DeleteOptions{}); err != nil {
		return multierr.Combine(errFailedToDeleteRole, err)
	}

	if err = c.clientset.RbacV1().Roles(namespaceCrossplane).Delete(ctx, roleCrossplane.Name, metav1.DeleteOptions{}); err != nil {
		return multierr.Combine(errFailedToDeleteRole, err)
	}

	if err = c.clientsetSA.Delete(ctx, serviceAccount.Name, metav1.DeleteOptions{}); err != nil {
		return multierr.Combine(errFailedToDeleteServiceAccount, err)
	}

	log.Println(logMsgResourcesCleanedUp)

	if pod.Status.Phase == corev1.PodFailed {
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

	if err = c.setupClientsets(); err != nil {
		log.Fatalln(err)
	}

	log.Println(logMsgKubeClientsetCreated)

	ctx := context.Background()

	serviceAccount, err := c.createServiceAccount(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	err = c.ensureNamespace(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	roleCrossplane, roleMySQL, err := c.createRoles(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	err = c.createRoleBindings(ctx, serviceAccount, roleCrossplane, roleMySQL)
	if err != nil {
		log.Fatalln(err)
	}

	err = c.createPod(ctx, serviceAccount)
	if err != nil {
		log.Fatalln(err)
	}

	err = c.waitForPodToRun(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	err = c.streamPodLogs(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	err = c.cleanupResources(ctx, roleCrossplane, roleMySQL, serviceAccount)
	if err != nil {
		log.Fatalln(err)
	}
}

// newCheckCmd returns a new Check command.
func newCheckCmd() *checkCmd {
	return &checkCmd{}
}

// Check returns a Cobra command for checking the infrastructure.
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
		defaultDockerRepo = "752320408524.dkr.ecr.us-east-1.amazonaws.com"
	)

	cobraCmd.Flags().String(
		flagKubeConfig,
		constant.EmptyString,
		"path to the Kubernetes configuration file to use for the check (or KUBECONFIG environment variable)",
	)

	cobraCmd.Flags().String(flagDockerRepo, defaultDockerRepo, "the Docker repository to use for the pod image")

	return cobraCmd
}
