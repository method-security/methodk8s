package ingress

import (
	"context"
	"fmt"
	"strings"

	methodk8s "github.com/method-security/methodk8s/generated/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func init() {
	// Register the Gateway API types with the scheme
	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme.Scheme))
}

func EnumerateIngresses(ctx context.Context, k8sconfig *rest.Config, authType methodk8s.AuthTypes, types []string) (*methodk8s.IngressReport, error) {
	resources := methodk8s.IngressReport{}
	errors := []string{}

	config := k8sconfig

	httpRoutes := []*methodk8s.HttpRoute{}
	if contains(types, "gateway") || len(types) == 0 {
		// Create a new Kubernetes client specific for gateways
		k8sClient, err := client.New(config, client.Options{Scheme: scheme.Scheme})
		if err != nil {
			return &methodk8s.IngressReport{}, err
		}

		httpRoutesList := &gatewayv1beta1.HTTPRouteList{}
		if err := k8sClient.List(ctx, httpRoutesList, &client.ListOptions{}); err != nil {
			errors = append(errors, err.Error())
		} else {
			for _, route := range httpRoutesList.Items {
				hostnames := []string{}
				for _, host := range route.Spec.Hostnames {
					hostnames = append(hostnames, string(host))
				}

				gateways := []*methodk8s.Gateway{}
				for _, gw := range route.Spec.ParentRefs {
					// Get the namespace of the gateway
					namespace := string(*gw.Namespace)

					gatewayInfo := &methodk8s.Gateway{
						Name:      string(gw.Name),
						Namespace: namespace,
					}
					gateways = append(gateways, gatewayInfo)
				}

				paths := []*methodk8s.Path{}
				for _, rule := range route.Spec.Rules {
					for _, match := range rule.Matches {
						if match.Path != nil && match.Path.Value != nil {
							// Iterate over BackendRefs to collect ports
							for _, backendRef := range rule.BackendRefs {
								for _, hostname := range hostnames {
									var namespace string
									if backendRef.Namespace != nil {
										namespace = string(*backendRef.Namespace)
									} else {
										namespace = route.GetNamespace()
									}

									pathInfo := &methodk8s.Path{
										Path:             *match.Path.Value,
										Base:             hostname,
										ServiceNamespace: namespace,
										ServiceName:      string(backendRef.Name),
									}
									if backendRef.Port != nil {
										port := fmt.Sprintf("%d", *backendRef.Port)
										pathInfo.Port = &port
									}
									paths = append(paths, pathInfo)
								}
							}
						}
					}
				}
				httpRouteInfo := &methodk8s.HttpRoute{
					Name:        route.GetName(),
					Namespace:   route.GetNamespace(),
					Labels:      route.GetLabels(),
					Annotations: route.GetAnnotations(),
					Gateways:    gateways,
					Paths:       paths,
				}
				httpRoutes = append(httpRoutes, httpRouteInfo)
			}
		}
	}
	ingresses := []*methodk8s.Ingress{}
	if contains(types, "ingress") || len(types) == 0 {
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return &methodk8s.IngressReport{}, err
		}
		ingressList, err := clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			for _, ingress := range ingressList.Items {

				rules := []*methodk8s.Rule{}
				for _, rule := range ingress.Spec.Rules {
					for _, path := range rule.HTTP.Paths {
						ruleInfo := methodk8s.Rule{
							Path:        path.Path,
							Base:        rule.Host,
							ServiceName: path.Backend.Service.Name,
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

	if len(httpRoutes) == 0 && len(ingresses) == 0 {
		return &methodk8s.IngressReport{AuthType: authType, Errors: errors}, nil
	}
	resources = methodk8s.IngressReport{
		HttpRoutes: httpRoutes,
		Ingresses:  ingresses,
		ClusterUrl: &config.Host,
		AuthType:   authType,
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
