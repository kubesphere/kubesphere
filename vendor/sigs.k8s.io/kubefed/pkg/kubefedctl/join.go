/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubefedctl

import (
	"context"
	goerrors "errors"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1b1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	genericclient "sigs.k8s.io/kubefed/pkg/client/generic"
	ctlutil "sigs.k8s.io/kubefed/pkg/controller/util"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/options"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

const (
	serviceAccountSecretTimeout = 30 * time.Second
)

var (
	join_long = `
		Join registers a Kubernetes cluster with a KubeFed control
		plane.

		Current context is assumed to be a Kubernetes cluster
		hosting a KubeFed control plane. Please use the
		--host-cluster-context flag otherwise.`
	join_example = `
		# Register a cluster with a KubeFed control plane by
		# specifying the cluster name and the context name of
		# the control plane's host cluster. Cluster name must
		# be a valid RFC 1123 subdomain name. Cluster context
		# must be specified if the cluster name is different
		# than the cluster's context in the local kubeconfig.
		kubefedctl join foo --host-cluster-context=bar`

	// Policy rules allowing full access to resources in the cluster
	// or namespace.
	namespacedPolicyRules = []rbacv1.PolicyRule{
		{
			Verbs:     []string{rbacv1.VerbAll},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{rbacv1.ResourceAll},
		},
	}
	clusterPolicyRules = []rbacv1.PolicyRule{
		namespacedPolicyRules[0],
		{
			NonResourceURLs: []string{rbacv1.NonResourceAll},
			Verbs:           []string{"get"},
		},
	}
)

type joinFederation struct {
	options.GlobalSubcommandOptions
	options.CommonJoinOptions
	joinFederationOptions
}

type joinFederationOptions struct {
	secretName      string
	scope           apiextv1b1.ResourceScope
	errorOnExisting bool
}

// Bind adds the join specific arguments to the flagset passed in as an
// argument.
func (o *joinFederationOptions) Bind(flags *pflag.FlagSet) {
	flags.StringVar(&o.secretName, "secret-name", "",
		"Name of the secret where the cluster's credentials will be stored in the host cluster. This name should be a valid RFC 1035 label. If unspecified, defaults to a generated name containing the cluster name.")
	flags.BoolVar(&o.errorOnExisting, "error-on-existing", true,
		"Whether the join operation will throw an error if it encounters existing artifacts with the same name as those it's trying to create. If false, the join operation will update existing artifacts to match its own specification.")
}

// NewCmdJoin defines the `join` command that registers a cluster with
// a KubeFed control plane.
func NewCmdJoin(cmdOut io.Writer, config util.FedConfig) *cobra.Command {
	opts := &joinFederation{}

	cmd := &cobra.Command{
		Use:     "join CLUSTER_NAME --host-cluster-context=HOST_CONTEXT",
		Short:   "Register a cluster with a KubeFed control plane",
		Long:    join_long,
		Example: join_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := opts.Complete(args)
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}

			err = opts.Run(cmdOut, config)
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}
		},
	}

	flags := cmd.Flags()
	opts.GlobalSubcommandBind(flags)
	opts.CommonSubcommandBind(flags)
	opts.Bind(flags)

	return cmd
}

// Complete ensures that options are valid and marshals them if necessary.
func (j *joinFederation) Complete(args []string) error {
	err := j.SetName(args)
	if err != nil {
		return err
	}

	if j.ClusterContext == "" {
		klog.V(2).Infof("Defaulting cluster context to joining cluster name %s", j.ClusterName)
		j.ClusterContext = j.ClusterName
	}

	if j.HostClusterName != "" && strings.ContainsAny(j.HostClusterName, ":/") {
		return goerrors.New("host-cluster-name may not contain \"/\" or \":\"")
	}

	if j.HostClusterName == "" && strings.ContainsAny(j.HostClusterContext, ":/") {
		klog.Fatal("host-cluster-name must be set if the name of the host cluster context contains one of \":\" or \"/\"")
	}

	klog.V(2).Infof("Args and flags: name %s, host: %s, host-system-namespace: %s, kubeconfig: %s, cluster-context: %s, secret-name: %s, dry-run: %v",
		j.ClusterName, j.HostClusterContext, j.KubeFedNamespace, j.Kubeconfig, j.ClusterContext,
		j.secretName, j.DryRun)

	return nil
}

// Run is the implementation of the `join` command.
func (j *joinFederation) Run(cmdOut io.Writer, config util.FedConfig) error {
	hostClientConfig := config.GetClientConfig(j.HostClusterContext, j.Kubeconfig)
	if err := j.SetHostClusterContextFromConfig(hostClientConfig); err != nil {
		return err
	}

	hostConfig, err := hostClientConfig.ClientConfig()
	if err != nil {
		// TODO(font): Return new error with this same text so it can be output
		// by caller.
		klog.V(2).Infof("Failed to get host cluster config: %v", err)
		return err
	}

	j.scope, err = options.GetScopeFromKubeFedConfig(hostConfig, j.KubeFedNamespace)
	if err != nil {
		return err
	}

	clusterConfig, err := config.ClusterConfig(j.ClusterContext, j.Kubeconfig)
	if err != nil {
		klog.V(2).Infof("Failed to get joining cluster config: %v", err)
		return err
	}

	hostClusterName := j.HostClusterContext
	if j.HostClusterName != "" {
		hostClusterName = j.HostClusterName
	}

	_, err = JoinCluster(hostConfig, clusterConfig, j.KubeFedNamespace,
		hostClusterName, j.ClusterName, j.secretName, j.scope, j.DryRun, j.errorOnExisting)

	return err
}

// JoinCluster registers a cluster with a KubeFed control plane. The
// KubeFed namespace in the joining cluster will be the same as in the
// host cluster.
func JoinCluster(hostConfig, clusterConfig *rest.Config, kubefedNamespace,
	hostClusterName, joiningClusterName, secretName string,
	scope apiextv1b1.ResourceScope, dryRun, errorOnExisting bool) (*fedv1b1.KubeFedCluster, error) {

	return joinClusterForNamespace(hostConfig, clusterConfig, kubefedNamespace,
		kubefedNamespace, hostClusterName, joiningClusterName, secretName,
		scope, dryRun, errorOnExisting)
}

// joinClusterForNamespace registers a cluster with a KubeFed control
// plane. The KubeFed namespace in the joining cluster is provided by
// the joiningNamespace parameter.
func joinClusterForNamespace(hostConfig, clusterConfig *rest.Config, kubefedNamespace,
	joiningNamespace, hostClusterName, joiningClusterName, secretName string,
	scope apiextv1b1.ResourceScope, dryRun, errorOnExisting bool) (*fedv1b1.KubeFedCluster, error) {

	hostClientset, err := util.HostClientset(hostConfig)
	if err != nil {
		klog.V(2).Infof("Failed to get host cluster clientset: %v", err)
		return nil, err
	}

	clusterClientset, err := util.ClusterClientset(clusterConfig)
	if err != nil {
		klog.V(2).Infof("Failed to get joining cluster clientset: %v", err)
		return nil, err
	}

	client, err := genericclient.New(hostConfig)
	if err != nil {
		klog.V(2).Infof("Failed to get kubefed clientset: %v", err)
		return nil, err
	}

	klog.V(2).Infof("Performing preflight checks.")
	err = performPreflightChecks(clusterClientset, joiningClusterName, hostClusterName, joiningNamespace, errorOnExisting)
	if err != nil {
		return nil, err
	}

	klog.V(2).Infof("Creating %s namespace in joining cluster", joiningNamespace)
	_, err = createKubeFedNamespace(clusterClientset, joiningNamespace,
		joiningClusterName, dryRun)
	if err != nil {
		klog.V(2).Infof("Error creating %s namespace in joining cluster: %v",
			joiningNamespace, err)
		return nil, err
	}
	klog.V(2).Infof("Created %s namespace in joining cluster", joiningNamespace)

	saName, err := createAuthorizedServiceAccount(clusterClientset,
		joiningNamespace, joiningClusterName, hostClusterName,
		scope, dryRun, errorOnExisting)
	if err != nil {
		return nil, err
	}

	secret, caBundle, err := populateSecretInHostCluster(clusterClientset, hostClientset,
		saName, kubefedNamespace, joiningNamespace, joiningClusterName, secretName, dryRun)
	if err != nil {
		klog.V(2).Infof("Error creating secret in host cluster: %s due to: %v", hostClusterName, err)
		return nil, err
	}

	var disabledTLSValidations []fedv1b1.TLSValidation
	if clusterConfig.TLSClientConfig.Insecure {
		disabledTLSValidations = append(disabledTLSValidations, fedv1b1.TLSAll)
	}

	kubefedCluster, err := createKubeFedCluster(client, joiningClusterName, clusterConfig.Host,
		secret.Name, kubefedNamespace, caBundle, disabledTLSValidations, dryRun, errorOnExisting)
	if err != nil {
		klog.V(2).Infof("Failed to create federated cluster resource: %v", err)
		return nil, err
	}

	klog.V(2).Info("Created federated cluster resource")
	return kubefedCluster, nil
}

// This function is exported for testing purposes only.
var TestOnly_JoinClusterForNamespace = joinClusterForNamespace

// performPreflightChecks checks that the host and joining clusters are in
// a consistent state.
func performPreflightChecks(clusterClientset kubeclient.Interface, name, hostClusterName,
	kubefedNamespace string, errorOnExisting bool) error {
	// Make sure there is no existing service account in the joining cluster.
	saName := util.ClusterServiceAccountName(name, hostClusterName)
	_, err := clusterClientset.CoreV1().ServiceAccounts(kubefedNamespace).Get(saName,
		metav1.GetOptions{})

	switch {
	case apierrors.IsNotFound(err):
		return nil
	case err != nil:
		return err
	case errorOnExisting:
		return errors.Errorf("service account: %s already exists in joining cluster: %s", saName, name)
	default:
		klog.V(2).Infof("Service account %s already exists in joining cluster %s", saName, name)
		return nil
	}
}

// createKubeFedCluster creates a federated cluster resource that associates
// the cluster and secret.
func createKubeFedCluster(client genericclient.Client, joiningClusterName, apiEndpoint,
	secretName, kubefedNamespace string, caBundle []byte, disabledTLSValidations []fedv1b1.TLSValidation,
	dryRun, errorOnExisting bool) (*fedv1b1.KubeFedCluster, error) {
	fedCluster := &fedv1b1.KubeFedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kubefedNamespace,
			Name:      joiningClusterName,
		},
		Spec: fedv1b1.KubeFedClusterSpec{
			APIEndpoint: apiEndpoint,
			CABundle:    caBundle,
			SecretRef: fedv1b1.LocalSecretReference{
				Name: secretName,
			},
			DisabledTLSValidations: disabledTLSValidations,
		},
	}

	if dryRun {
		return fedCluster, nil
	}

	existingFedCluster := &fedv1b1.KubeFedCluster{}
	err := client.Get(context.TODO(), existingFedCluster, kubefedNamespace, joiningClusterName)
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not retrieve federated cluster %s due to %v", joiningClusterName, err)
		return nil, err
	case err == nil && errorOnExisting:
		return nil, errors.Errorf("federated cluster %s already exists in host cluster", joiningClusterName)
	case err == nil:
		existingFedCluster.Spec = fedCluster.Spec
		err = client.Update(context.TODO(), existingFedCluster)
		if err != nil {
			klog.V(2).Infof("Could not update federated cluster %s due to %v", fedCluster.Name, err)
			return nil, err
		}
		return existingFedCluster, nil
	default:
		err = client.Create(context.TODO(), fedCluster)
		if err != nil {
			klog.V(2).Infof("Could not create federated cluster %s due to %v", fedCluster.Name, err)
			return nil, err
		}
		return fedCluster, nil
	}
}

// createKubeFedNamespace creates the kubefed namespace in the cluster
// associated with clusterClientset, if it doesn't already exist.
func createKubeFedNamespace(clusterClientset kubeclient.Interface, kubefedNamespace,
	joiningClusterName string, dryRun bool) (*corev1.Namespace, error) {
	fedNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubefedNamespace,
		},
	}

	if dryRun {
		return fedNamespace, nil
	}

	_, err := clusterClientset.CoreV1().Namespaces().Get(kubefedNamespace, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.V(2).Infof("Could not get %s namespace: %v", kubefedNamespace, err)
		return nil, err
	}

	if err == nil {
		klog.V(2).Infof("Already existing %s namespace", kubefedNamespace)
		return fedNamespace, nil
	}

	// Not found, so create.
	_, err = clusterClientset.CoreV1().Namespaces().Create(fedNamespace)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		klog.V(2).Infof("Could not create %s namespace: %v", kubefedNamespace, err)
		return nil, err
	}
	return fedNamespace, nil
}

// createAuthorizedServiceAccount creates a service account and grants
// the privileges required by the KubeFed control plane to manage
// resources in the joining cluster.  The name of the created service
// account is returned on success.
func createAuthorizedServiceAccount(joiningClusterClientset kubeclient.Interface,
	namespace, joiningClusterName, hostClusterName string,
	scope apiextv1b1.ResourceScope, dryRun, errorOnExisting bool) (string, error) {

	klog.V(2).Infof("Creating service account in joining cluster: %s", joiningClusterName)

	saName, err := createServiceAccount(joiningClusterClientset, namespace,
		joiningClusterName, hostClusterName, dryRun, errorOnExisting)
	if err != nil {
		klog.V(2).Infof("Error creating service account: %s in joining cluster: %s due to: %v",
			saName, joiningClusterName, err)
		return "", err
	}

	klog.V(2).Infof("Created service account: %s in joining cluster: %s", saName, joiningClusterName)

	if scope == apiextv1b1.NamespaceScoped {
		klog.V(2).Infof("Creating role and binding for service account: %s in joining cluster: %s", saName, joiningClusterName)

		err = createRoleAndBinding(joiningClusterClientset, saName, namespace, joiningClusterName, dryRun, errorOnExisting)
		if err != nil {
			klog.V(2).Infof("Error creating role and binding for service account: %s in joining cluster: %s due to: %v", saName, joiningClusterName, err)
			return "", err
		}

		klog.V(2).Infof("Created role and binding for service account: %s in joining cluster: %s",
			saName, joiningClusterName)

		klog.V(2).Infof("Creating health check cluster role and binding for service account: %s in joining cluster: %s", saName, joiningClusterName)

		err = createHealthCheckClusterRoleAndBinding(joiningClusterClientset, saName, namespace, joiningClusterName,
			dryRun, errorOnExisting)
		if err != nil {
			klog.V(2).Infof("Error creating health check cluster role and binding for service account: %s in joining cluster: %s due to: %v",
				saName, joiningClusterName, err)
			return "", err
		}

		klog.V(2).Infof("Created health check cluster role and binding for service account: %s in joining cluster: %s",
			saName, joiningClusterName)

	} else {
		klog.V(2).Infof("Creating cluster role and binding for service account: %s in joining cluster: %s", saName, joiningClusterName)

		err = createClusterRoleAndBinding(joiningClusterClientset, saName, namespace, joiningClusterName, dryRun, errorOnExisting)
		if err != nil {
			klog.V(2).Infof("Error creating cluster role and binding for service account: %s in joining cluster: %s due to: %v",
				saName, joiningClusterName, err)
			return "", err
		}

		klog.V(2).Infof("Created cluster role and binding for service account: %s in joining cluster: %s",
			saName, joiningClusterName)
	}

	return saName, nil
}

// createServiceAccount creates a service account in the cluster associated
// with clusterClientset with credentials that will be used by the host cluster
// to access its API server.
func createServiceAccount(clusterClientset kubeclient.Interface, namespace,
	joiningClusterName, hostClusterName string, dryRun, errorOnExisting bool) (string, error) {
	saName := util.ClusterServiceAccountName(joiningClusterName, hostClusterName)
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: namespace,
		},
	}

	if dryRun {
		return saName, nil
	}

	// Create a new service account.
	_, err := clusterClientset.CoreV1().ServiceAccounts(namespace).Create(sa)
	switch {
	case apierrors.IsAlreadyExists(err) && errorOnExisting:
		klog.V(2).Infof("Service account %s/%s already exists in target cluster %s", namespace, saName, joiningClusterName)
		return "", err
	case err != nil && !apierrors.IsAlreadyExists(err):
		klog.V(2).Infof("Could not create service account %s/%s in target cluster %s due to: %v", namespace, saName, joiningClusterName, err)
		return "", err
	default:
		return saName, nil
	}
}

func bindingSubjects(saName, namespace string) []rbacv1.Subject {
	return []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      saName,
			Namespace: namespace,
		},
	}
}

// createClusterRoleAndBinding creates an RBAC cluster role and
// binding that allows the service account identified by saName to
// access all resources in all namespaces in the cluster associated
// with clientset.
func createClusterRoleAndBinding(clientset kubeclient.Interface, saName, namespace, clusterName string, dryRun, errorOnExisting bool) error {
	if dryRun {
		return nil
	}

	roleName := util.RoleName(saName)

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Rules: clusterPolicyRules,
	}
	existingRole, err := clientset.RbacV1().ClusterRoles().Get(roleName, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not get cluster role for service account %s in joining cluster %s due to %v",
			saName, clusterName, err)
		return err
	case err == nil && errorOnExisting:
		return errors.Errorf("cluster role for service account %s in joining cluster %s already exists", saName, clusterName)
	case err == nil:
		existingRole.Rules = role.Rules
		_, err := clientset.RbacV1().ClusterRoles().Update(existingRole)
		if err != nil {
			klog.V(2).Infof("Could not update cluster role for service account: %s in joining cluster: %s due to: %v",
				saName, clusterName, err)
			return err
		}
	default: // role was not found
		_, err := clientset.RbacV1().ClusterRoles().Create(role)
		if err != nil {
			klog.V(2).Infof("Could not create cluster role for service account: %s in joining cluster: %s due to: %v",
				saName, clusterName, err)
			return err
		}
	}

	// TODO: This should limit its access to only necessary resources.
	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Subjects: bindingSubjects(saName, namespace),
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     roleName,
		},
	}
	existingBinding, err := clientset.RbacV1().ClusterRoleBindings().Get(binding.Name, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not get cluster role binding for service account %s in joining cluster %s due to %v",
			saName, clusterName, err)
		return err
	case err == nil && errorOnExisting:
		return errors.Errorf("cluster role binding for service account %s in joining cluster %s already exists", saName, clusterName)
	case err == nil:
		// The roleRef cannot be updated, therefore if the existing roleRef is different, the existing rolebinding
		// must be deleted and recreated with the correct roleRef
		if !reflect.DeepEqual(existingBinding.RoleRef, binding.RoleRef) {
			err = clientset.RbacV1().ClusterRoleBindings().Delete(existingBinding.Name, &metav1.DeleteOptions{})
			if err != nil {
				klog.V(2).Infof("Could not delete existing cluster role binding for service account %s in joining cluster %s due to: %v",
					saName, clusterName, err)
				return err
			}
			_, err = clientset.RbacV1().ClusterRoleBindings().Create(binding)
			if err != nil {
				klog.V(2).Infof("Could not create cluster role binding for service account: %s in joining cluster: %s due to: %v",
					saName, clusterName, err)
				return err
			}
		} else {
			existingBinding.Subjects = binding.Subjects
			_, err := clientset.RbacV1().ClusterRoleBindings().Update(existingBinding)
			if err != nil {
				klog.V(2).Infof("Could not update cluster role binding for service account: %s in joining cluster: %s due to: %v",
					saName, clusterName, err)
				return err
			}
		}
	default:
		_, err = clientset.RbacV1().ClusterRoleBindings().Create(binding)
		if err != nil {
			klog.V(2).Infof("Could not create cluster role binding for service account: %s in joining cluster: %s due to: %v",
				saName, clusterName, err)
			return err
		}
	}
	return nil
}

// createRoleAndBinding creates an RBAC role and binding
// that allows the service account identified by saName to access all
// resources in the specified namespace.
func createRoleAndBinding(clientset kubeclient.Interface, saName, namespace, clusterName string, dryRun, errorOnExisting bool) error {
	if dryRun {
		return nil
	}

	roleName := util.RoleName(saName)

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Rules: namespacedPolicyRules,
	}
	existingRole, err := clientset.RbacV1().Roles(namespace).Get(roleName, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not retrieve role for service account %s in joining cluster %s due to %v", saName, clusterName, err)
		return err
	case errorOnExisting && err == nil:
		return errors.Errorf("role for service account %s in joining cluster %s already exists", saName, clusterName)
	case err == nil:
		existingRole.Rules = role.Rules
		_, err = clientset.RbacV1().Roles(namespace).Update(existingRole)
		if err != nil {
			klog.V(2).Infof("Could not update role for service account: %s in joining cluster: %s due to: %v",
				saName, clusterName, err)
			return err
		}
	default:
		_, err := clientset.RbacV1().Roles(namespace).Create(role)
		if err != nil {
			klog.V(2).Infof("Could not create role for service account: %s in joining cluster: %s due to: %v",
				saName, clusterName, err)
			return err
		}
	}

	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Subjects: bindingSubjects(saName, namespace),
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     roleName,
		},
	}

	existingBinding, err := clientset.RbacV1().RoleBindings(namespace).Get(binding.Name, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not retrieve role binding for service account %s in joining cluster %s due to: %v",
			saName, clusterName, err)
		return err
	case err == nil && errorOnExisting:
		return errors.Errorf("role binding for service account %s in joining cluster %s already exists", saName, clusterName)
	case err == nil:
		// The roleRef cannot be updated, therefore if the existing roleRef is different, the existing rolebinding
		// must be deleted and recreated with the correct roleRef
		if !reflect.DeepEqual(existingBinding.RoleRef, binding.RoleRef) {
			err = clientset.RbacV1().RoleBindings(namespace).Delete(existingBinding.Name, &metav1.DeleteOptions{})
			if err != nil {
				klog.V(2).Infof("Could not delete existing role binding for service account %s in joining cluster %s due to: %v",
					saName, clusterName, err)
				return err
			}
			_, err = clientset.RbacV1().RoleBindings(namespace).Create(binding)
			if err != nil {
				klog.V(2).Infof("Could not create role binding for service account: %s in joining cluster: %s due to: %v",
					saName, clusterName, err)
				return err
			}
		} else {
			existingBinding.Subjects = binding.Subjects
			_, err = clientset.RbacV1().RoleBindings(namespace).Update(existingBinding)
			if err != nil {
				klog.V(2).Infof("Could not update role binding for service account %s in joining cluster %s due to: %v",
					saName, clusterName, err)
				return err
			}
		}
	default:
		_, err = clientset.RbacV1().RoleBindings(namespace).Create(binding)
		if err != nil {
			klog.V(2).Infof("Could not create role binding for service account: %s in joining cluster: %s due to: %v",
				saName, clusterName, err)
			return err
		}
	}

	return nil
}

// createHealthCheckClusterRoleAndBinding creates an RBAC cluster role and
// binding that allows the service account identified by saName to
// access the health check path of the cluster.
func createHealthCheckClusterRoleAndBinding(clientset kubeclient.Interface, saName, namespace, clusterName string, dryRun, errorOnExisting bool) error {
	if dryRun {
		return nil
	}

	roleName := util.HealthCheckRoleName(saName, namespace)

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:           []string{"Get"},
				NonResourceURLs: []string{"/healthz"},
			},
			// The cluster client expects to be able to list nodes to retrieve zone and region details.
			// TODO(marun) Consider making zone/region retrieval optional
			{
				Verbs:     []string{"list"},
				APIGroups: []string{""},
				Resources: []string{"nodes"},
			},
		},
	}
	existingRole, err := clientset.RbacV1().ClusterRoles().Get(role.Name, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not get health check cluster role for service account %s in joining cluster %s due to %v",
			saName, clusterName, err)
		return err
	case err == nil && errorOnExisting:
		return errors.Errorf("health check cluster role for service account %s in joining cluster %s already exists", saName, clusterName)
	case err == nil:
		existingRole.Rules = role.Rules
		_, err := clientset.RbacV1().ClusterRoles().Update(existingRole)
		if err != nil {
			klog.V(2).Infof("Could not update health check cluster role for service account: %s in joining cluster: %s due to: %v",
				saName, clusterName, err)
			return err
		}
	default: // role was not found
		_, err := clientset.RbacV1().ClusterRoles().Create(role)
		if err != nil {
			klog.V(2).Infof("Could not create health check cluster role for service account: %s in joining cluster: %s due to: %v",
				saName, clusterName, err)
			return err
		}
	}

	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Subjects: bindingSubjects(saName, namespace),
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     roleName,
		},
	}
	existingBinding, err := clientset.RbacV1().ClusterRoleBindings().Get(binding.Name, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not get health check cluster role binding for service account %s in joining cluster %s due to %v",
			saName, clusterName, err)
		return err
	case err == nil && errorOnExisting:
		return errors.Errorf("health check cluster role binding for service account %s in joining cluster %s already exists", saName, clusterName)
	case err == nil:
		// The roleRef cannot be updated, therefore if the existing roleRef is different, the existing rolebinding
		// must be deleted and recreated with the correct roleRef
		if !reflect.DeepEqual(existingBinding.RoleRef, binding.RoleRef) {
			err = clientset.RbacV1().ClusterRoleBindings().Delete(existingBinding.Name, &metav1.DeleteOptions{})
			if err != nil {
				klog.V(2).Infof("Could not delete existing health check cluster role binding for service account %s in joining cluster %s due to: %v",
					saName, clusterName, err)
				return err
			}
			_, err = clientset.RbacV1().ClusterRoleBindings().Create(binding)
			if err != nil {
				klog.V(2).Infof("Could not create health check cluster role binding for service account: %s in joining cluster: %s due to: %v",
					saName, clusterName, err)
				return err
			}
		} else {
			existingBinding.Subjects = binding.Subjects
			_, err := clientset.RbacV1().ClusterRoleBindings().Update(existingBinding)
			if err != nil {
				klog.V(2).Infof("Could not update health check cluster role binding for service account: %s in joining cluster: %s due to: %v",
					saName, clusterName, err)
				return err
			}
		}
	default:
		_, err = clientset.RbacV1().ClusterRoleBindings().Create(binding)
		if err != nil {
			klog.V(2).Infof("Could not create health check cluster role binding for service account: %s in joining cluster: %s due to: %v",
				saName, clusterName, err)
			return err
		}
	}
	return nil
}

// populateSecretInHostCluster copies the service account secret for saName
// from the cluster referenced by clusterClientset to the client referenced by
// hostClientset, putting it in a secret named secretName in the provided
// namespace.
func populateSecretInHostCluster(clusterClientset, hostClientset kubeclient.Interface,
	saName, hostNamespace, joiningNamespace, joiningClusterName, secretName string,
	dryRun bool) (*corev1.Secret, []byte, error) {

	klog.V(2).Infof("Creating cluster credentials secret in host cluster")

	if dryRun {
		dryRunSecret := &corev1.Secret{}
		dryRunSecret.Name = secretName
		return dryRunSecret, nil, nil
	}

	// Get the secret from the joining cluster.
	var secret *corev1.Secret
	err := wait.PollImmediate(1*time.Second, serviceAccountSecretTimeout, func() (bool, error) {
		sa, err := clusterClientset.CoreV1().ServiceAccounts(joiningNamespace).Get(saName,
			metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, objReference := range sa.Secrets {
			saSecretName := objReference.Name
			var err error
			secret, err = clusterClientset.CoreV1().Secrets(joiningNamespace).Get(saSecretName,
				metav1.GetOptions{})
			if err != nil {
				return false, nil
			}
			if secret.Type == corev1.SecretTypeServiceAccountToken {
				klog.V(2).Infof("Using secret named: %s", secret.Name)
				return true, nil
			}
		}
		return false, nil
	})

	if err != nil {
		klog.V(2).Infof("Could not get service account secret from joining cluster: %v", err)
		return nil, nil, err
	}

	token, ok := secret.Data[ctlutil.TokenKey]
	if !ok {
		return nil, nil, errors.Errorf("Key %q not found in service account secret", ctlutil.TokenKey)
	}

	// Create a secret in the host cluster containing the token.
	v1Secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hostNamespace,
		},
		Data: map[string][]byte{
			ctlutil.TokenKey: token,
		},
	}

	if secretName == "" {
		v1Secret.GenerateName = joiningClusterName + "-"
	} else {
		v1Secret.Name = secretName
	}

	v1SecretResult, err := hostClientset.CoreV1().Secrets(hostNamespace).Create(&v1Secret)
	if err != nil {
		klog.V(2).Infof("Could not create secret in host cluster: %v", err)
		return nil, nil, err
	}

	// caBundle is optional so no error is suggested if it is not
	// found in the secret.
	caBundle := secret.Data["ca.crt"]

	klog.V(2).Infof("Created secret in host cluster named: %s", v1SecretResult.Name)
	return v1SecretResult, caBundle, nil
}
