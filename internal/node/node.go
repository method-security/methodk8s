package node

import (
	"context"
	"regexp"
	"strings"

	methodk8s "github.com/method-security/methodk8s/generated/go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func EnumerateNodes(ctx context.Context, k8sconfig *rest.Config, authType methodk8s.AuthTypes) (*methodk8s.NodeReport, error) {
	resources := methodk8s.NodeReport{}
	errors := []string{}

	config := k8sconfig

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &methodk8s.NodeReport{}, err
	}

	// Fetch all Pods
	nodesList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		errors = append(errors, err.Error())
		return &methodk8s.NodeReport{AuthType: authType, Errors: errors}, nil
	}

	// Loop through Nodes
	nodes := []*methodk8s.Node{}
	for _, node := range nodesList.Items {
		addresses := []*methodk8s.AddressInfo{}
		for _, addr := range node.Status.Addresses {
			addressType := string(addr.Type)
			addressInfo := methodk8s.AddressInfo{
				Type:    addressType,
				Address: addr.Address,
			}
			addresses = append(addresses, &addressInfo)
		}

		instanceType := node.Labels["node.kubernetes.io/instance-type"]
		nodeState, _ := whatState(&node)

		image, version := splitOnFirstNumber(node.Status.NodeInfo.OSImage)
		nodeInfo := methodk8s.Node{
			Name:         node.GetName(),
			Arch:         node.Status.NodeInfo.Architecture,
			Os:           node.Status.NodeInfo.OperatingSystem,
			Image:        image,
			Version:      version,
			Instancetype: &instanceType,
			State:        nodeState,
			Addresses:    addresses,
		}
		nodes = append(nodes, &nodeInfo)
	}

	resources = methodk8s.NodeReport{
		Nodes:      nodes,
		ClusterUrl: &config.Host,
		AuthType:   authType,
		Errors:     errors,
	}

	return &resources, nil
}

func whatState(node *corev1.Node) (methodk8s.StateTypes, error) {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return methodk8s.NewStateTypesFromString("Running")
		}
	}
	return methodk8s.NewStateTypesFromString("Stopped")
}

func splitOnFirstNumber(input string) (string, *string) {
	re := regexp.MustCompile(`\d`)
	loc := re.FindStringIndex(input)
	if loc != nil {
		part1 := strings.TrimSpace(input[:loc[0]]) // Remove trailing spaces
		part2 := input[loc[0]:]
		return part1, &part2
	}
	return strings.TrimSpace(input), nil
}
