package serviceaccount

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

func getClusterRoleBinding(namespace, serviceAccountName string) *rbacv1.ClusterRoleBinding {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "method-read-only-binding",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "method-read-only-clusterrole",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
	}
	return clusterRoleBinding
}

func getSecret(namespace, serviceAccountName string) *corev1.Secret {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "method-sa-secret",
			Namespace: namespace,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": serviceAccountName,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}
	return secret
}

func getServiceAccount(namespace, serviceAccountName string) *corev1.ServiceAccount {
	serviceAccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}
	return serviceAccount
}

func getClusterRole() *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "method-read-only-clusterrole",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints", "nodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "daemonsets", "replicasets", "statefulsets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs", "cronjobs"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"gateway.networking.k8s.io"},
				Resources: []string{"gateways"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	return clusterRole
}

func Config(ctx context.Context, k8sconfig *rest.Config, run bool, namespace string) error {
	config := k8sconfig
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	serviceAccountName := "method-service-account"
	k8sObjects := []interface{}{}

	clusterRole := getClusterRole()
	k8sObjects = append(k8sObjects, clusterRole)

	serviceAccount := getServiceAccount(namespace, serviceAccountName)
	k8sObjects = append(k8sObjects, serviceAccount)

	secret := getSecret(namespace, serviceAccountName)
	k8sObjects = append(k8sObjects, secret)

	clusterRoleBinding := getClusterRoleBinding(namespace, serviceAccountName)
	k8sObjects = append(k8sObjects, clusterRoleBinding)

	for _, obj := range k8sObjects {
		if run {
			switch o := obj.(type) {
			case *rbacv1.ClusterRole:
				_, err := clientset.RbacV1().ClusterRoles().Create(ctx, o, metav1.CreateOptions{})
				if err != nil && !errors.IsAlreadyExists(err) {
					return err
				}
				fmt.Println("- ClusterRole configured (1/4)")
			case *corev1.ServiceAccount:
				_, err := clientset.CoreV1().ServiceAccounts(namespace).Create(ctx, o, metav1.CreateOptions{})
				if err != nil && !errors.IsAlreadyExists(err) {
					return err
				}
				fmt.Println("- ServiceAccount configured (2/4)")
			case *corev1.Secret:
				_, err := clientset.CoreV1().Secrets(namespace).Create(ctx, o, metav1.CreateOptions{})
				if err != nil && !errors.IsAlreadyExists(err) {
					return err
				}
				fmt.Println("- Secret configured (3/4))")
			case *rbacv1.ClusterRoleBinding:
				_, err := clientset.RbacV1().ClusterRoleBindings().Create(ctx, o, metav1.CreateOptions{})
				if err != nil && !errors.IsAlreadyExists(err) {
					return err
				}
				fmt.Println("- ClusterRoleBinding configured (4/4)")
			}
		} else {
			fmt.Println("---")
			prettyPrintYAML(obj)
		}
	}
	return nil
}

// prettyPrintYAML prints a Kubernetes object as pretty YAML to the console
func prettyPrintYAML(obj interface{}) {
	yamlData, err := yaml.Marshal(obj)
	if err != nil {
		fmt.Printf("Error marshalling to YAML: %v\n", err)
		return
	}
	fmt.Printf("%s\n", string(yamlData))
}
