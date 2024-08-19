package service

import (
	"context"

	methodk8s "github.com/method-security/methodk8s/generated/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func EnumerateServices(ctx context.Context, k8sconfig *rest.Config, authType methodk8s.AuthTypes) (*methodk8s.ServiceReport, error) {
	resources := methodk8s.ServiceReport{}
	errors := []string{}

	config := k8sconfig

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &methodk8s.ServiceReport{}, err
	}

	servicesList, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errors = append(errors, err.Error())
		return &methodk8s.ServiceReport{AuthType: authType, Errors: errors}, nil
	}

	services := []*methodk8s.Service{}
	for _, service := range servicesList.Items {
		// Get the related Pod UIDs
		podUIDs, err := getPodUIDsForService(ctx, clientset, service.GetNamespace(), service.Spec.Selector)
		if err != nil {
			errors = append(errors, err.Error())
		}

		serviceInfo := methodk8s.Service{
			Name:        service.GetName(),
			Namespace:   service.GetNamespace(),
			Type:        string(service.Spec.Type),
			Pods:        podUIDs,
			Annotations: service.Annotations,
			Labels:      service.Labels,
			Selectors:   service.Spec.Selector,
		}
		services = append(services, &serviceInfo)
	}

	resources = methodk8s.ServiceReport{
		Services:   services,
		ClusterUrl: &config.Host,
		AuthType:   authType,
		Errors:     errors,
	}

	return &resources, nil
}

// getPodUIDsForService returns the UIDs of all pods related to the service based on the service's selectors
func getPodUIDsForService(ctx context.Context, clientset *kubernetes.Clientset, namespace string, selectors map[string]string) ([]string, error) {
	podUIDs := []string{}

	// Convert selectors map to a label selector string
	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: selectors})
	podsList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	for _, pod := range podsList.Items {
		podUIDs = append(podUIDs, string(pod.UID))
	}

	return podUIDs, nil
}
