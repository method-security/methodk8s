package pod

import (
	"context"

	methodk8s "github.com/method-security/methodk8s/generated/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func EnumeratePods(ctx context.Context, k8sconfig *rest.Config, authType methodk8s.AuthTypes) (*methodk8s.PodReport, error) {
	resources := methodk8s.PodReport{}
	errors := []string{}

	config := k8sconfig

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &methodk8s.PodReport{}, err
	}

	podsList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errors = append(errors, err.Error())
		return &methodk8s.PodReport{AuthType: authType, Errors: errors}, nil
	}

	pods := []*methodk8s.Pod{}
	for _, pod := range podsList.Items {
		namespace, err := clientset.CoreV1().Namespaces().Get(ctx, pod.GetNamespace(), metav1.GetOptions{})
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		containers := []*methodk8s.Container{}
		for _, container := range pod.Spec.Containers {
			ports := []*methodk8s.ContainerPort{}
			for _, port := range container.Ports {
				protocol, err := methodk8s.NewProtocolTypesFromString(string(port.Protocol))
				if err != nil {
					errors = append(errors, err.Error())
					protocol, _ = methodk8s.NewProtocolTypesFromString("UNDEFINED")
				}
				portInfo := methodk8s.ContainerPort{
					Port:     int(port.ContainerPort),
					Protocol: protocol,
				}
				ports = append(ports, &portInfo)
			}

			runAsRoot := container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil && *container.SecurityContext.RunAsUser == 0
			allowPrivilegeEscalation := container.SecurityContext != nil && container.SecurityContext.AllowPrivilegeEscalation != nil && *container.SecurityContext.AllowPrivilegeEscalation
			readOnlyRootFilesystem := container.SecurityContext != nil && container.SecurityContext.ReadOnlyRootFilesystem != nil && *container.SecurityContext.ReadOnlyRootFilesystem

			securityContext := methodk8s.SecurityContext{
				RunAsRoot:                &runAsRoot,
				AllowPrivilegeEscalation: &allowPrivilegeEscalation,
				ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
			}
			containerInfo := methodk8s.Container{
				Name:            container.Name,
				Image:           container.Image,
				Ports:           ports,
				SecurityContext: &securityContext,
			}
			containers = append(containers, &containerInfo)
		}

		status, err := methodk8s.NewStatusTypesFromString(string(pod.Status.Phase))
		if err != nil {
			errors = append(errors, err.Error())
			status, _ = methodk8s.NewStatusTypesFromString("Unknown")
		}

		namespaceInfo := methodk8s.NamespaceInfo{
			Name: namespace.GetName(),
			Uid:  string(namespace.GetUID()),
		}
		statusInfo := methodk8s.Status{
			Status: status,
			PodIp:  &pod.Status.PodIP,
			HostIp: &pod.Status.HostIP,
		}
		version := pod.GetResourceVersion()
		podInfo := methodk8s.Pod{
			Uid:         string(pod.GetUID()),
			Name:        pod.GetName(),
			Namespace:   &namespaceInfo,
			Status:      &statusInfo,
			Version:     &version,
			Node:        pod.Spec.NodeName,
			Containers:  containers,
			Annotations: pod.Annotations,
			Labels:      pod.Labels,
		}
		pods = append(pods, &podInfo)
	}

	resources = methodk8s.PodReport{
		Pods:       pods,
		ClusterUrl: &config.Host,
		AuthType:   authType,
		Errors:     errors,
	}
	return &resources, nil
}
