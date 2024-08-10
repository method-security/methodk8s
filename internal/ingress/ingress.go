package ingress

import (
	"context"
	"fmt"
	"strings"

	methodk8s "github.com/method-security/methodk8s/generated/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	gatewayclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

func EnumerateIngresses(ctx context.Context, k8config *rest.Config, types []string) (*methodk8s.IngressReport, error) {
	resources := methodk8s.IngressReport{}
	errors := []string{}

	clientset, err := kubernetes.NewForConfig(k8config)
	if err != nil {
		return &methodk8s.IngressReport{Errors: errors}, err
	}

	gatewayClient, err := gatewayclientset.NewForConfig(k8config)
	if err != nil {
		return &methodk8s.IngressReport{Errors: errors}, err
	}

	gateways := []*methodk8s.Gateway{}
	if contains(types, "gateway") || len(types) == 0 {
		gatewayList, err := gatewayClient.GatewayV1beta1().Gateways("").List(ctx, metav1.ListOptions{})
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			for _, gateway := range gatewayList.Items {
				listeners := []string{}
				for _, listener := range gateway.Spec.Listeners {
					var gatewayURL string

					for _, gateAddress := range gateway.Status.Addresses {
						address := gateAddress.Value
						gatewayURL = fmt.Sprintf("%s:%d", address, listener.Port)
						listeners = append(listeners, gatewayURL)
					}
				}

				gatewayInfo := &methodk8s.Gateway{
					Name:        gateway.GetName(),
					Namespace:   gateway.GetNamespace(),
					Listeners:   listeners,
					Annotations: gateway.GetAnnotations(),
					Labels:      gateway.GetLabels(),
				}
				gateways = append(gateways, gatewayInfo)
			}
		}
	}

	ingresses := []*methodk8s.Ingress{}
	if contains(types, "ingress") || len(types) == 0 {
		ingressList, err := clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			for _, ingress := range ingressList.Items {
				rules := []*methodk8s.Rule{}

				// Extract Hosts and Paths
				for _, rule := range ingress.Spec.Rules {
					for _, path := range rule.HTTP.Paths {

						var port *string
						if path.Backend.Service != nil && path.Backend.Service.Port.Number != 0 {
							portStr := fmt.Sprintf("%d", path.Backend.Service.Port.Number)
							port = &portStr
						} else {
							port = nil
						}
						ruleInfo := methodk8s.Rule{
							Host:        rule.Host,
							Path:        path.Path,
							ServiceName: path.Backend.Service.Name,
							ServicePort: port,
						}
						rules = append(rules, &ruleInfo)
					}
				}

				ingressInfo := &methodk8s.Ingress{
					Name:        ingress.GetName(),
					Namespace:   ingress.GetNamespace(),
					Rules:       rules,
					Annotations: ingress.GetAnnotations(),
					Labels:      ingress.GetLabels(),
				}
				ingresses = append(ingresses, ingressInfo)
			}
		}
	}

	resources = methodk8s.IngressReport{
		Gateways:   gateways,
		Ingresses:  ingresses,
		ClusterUrl: &k8config.Host,
		Errors:     errors,
	}

	return &resources, nil
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, item) {
			return true
		}
	}
	return false
}
