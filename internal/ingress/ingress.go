package ingress

import (
	"context"
	"fmt"
	"strings"

	methodk8s "github.com/method-security/methodk8s/generated/go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func init() {
	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme.Scheme))
}

func EnumerateIngresses(ctx context.Context, k8sconfig *rest.Config, authType methodk8s.AuthTypes, types []string) (*methodk8s.IngressReport, error) {
	resources := methodk8s.IngressReport{}
	errors := []string{}

	config := k8sconfig

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &methodk8s.IngressReport{}, err
	}

	// HttpRoutes
	httpRoutes := []*methodk8s.HttpRoute{}
	if contains(types, "gateway") || len(types) == 0 {
		k8sClient, err := client.New(config, client.Options{Scheme: scheme.Scheme})
		if err != nil {
			return &methodk8s.IngressReport{}, err
		}
		// Loop through HttpRoutes
		httpRoutesList := &gatewayv1beta1.HTTPRouteList{}
		if err := k8sClient.List(ctx, httpRoutesList, &client.ListOptions{}); err != nil {
			errors = append(errors, err.Error())
		} else {
			for _, route := range httpRoutesList.Items {
				// HttpRoute Namespace
				namespace, err := clientset.CoreV1().Namespaces().Get(ctx, route.GetNamespace(), metav1.GetOptions{})
				if err != nil {
					errors = append(errors, err.Error())
					continue
				}
				namespaceInfo := methodk8s.NamespaceInfo{
					Name: namespace.GetName(),
					Uid:  string(namespace.GetUID()),
				}
				// Route bases (ie. xxx.ec2.com)
				hostnames := []string{}
				for _, host := range route.Spec.Hostnames {
					hostnames = append(hostnames, string(host))
				}
				// Loop through Gateways that use these defined routes
				gateways := []*methodk8s.GatewayInfo{}
				for _, gw := range route.Spec.ParentRefs {
					// Gateway Namespace
					gatewayNamespace, err := clientset.CoreV1().Namespaces().Get(ctx, string(*gw.Namespace), metav1.GetOptions{})
					if err != nil {
						errors = append(errors, err.Error())
						continue
					}
					// Gateway object
					gateway := &gatewayv1beta1.Gateway{}
					err = k8sClient.Get(ctx, client.ObjectKey{
						Namespace: gatewayNamespace.GetName(),
						Name:      string(gw.Name),
					}, gateway)
					if err != nil {
						errors = append(errors, err.Error())
						continue
					}
					// Gateway data
					gatewayInfo := &methodk8s.GatewayInfo{
						Name: string(gw.Name),
						Uid:  string(gateway.GetUID()),
						Namespace: &methodk8s.NamespaceInfo{
							Name: gatewayNamespace.GetName(),
							Uid:  string(gatewayNamespace.GetUID()),
						},
					}
					gateways = append(gateways, gatewayInfo)
				}

				// Loop through HttpRoute routes
				paths := []*methodk8s.RouteInfo{}
				for _, rule := range route.Spec.Rules {
					for _, match := range rule.Matches {
						if match.Path != nil && match.Path.Value != nil {
							for _, backendRef := range rule.BackendRefs {
								for _, hostname := range hostnames {
									var backendNamespace string
									if backendRef.Namespace != nil {
										backendNamespace = string(*backendRef.Namespace)
									} else {
										backendNamespace = route.GetNamespace()
									}
									// Service Object
									service, err := clientset.CoreV1().Services(backendNamespace).Get(ctx, string(backendRef.Name), metav1.GetOptions{})
									if err != nil {
										errors = append(errors, err.Error())
										continue
									}
									// Service Namespace
									serviceNamespace, err := clientset.CoreV1().Namespaces().Get(ctx, backendNamespace, metav1.GetOptions{})
									if err != nil {
										errors = append(errors, err.Error())
										continue
									}
									// Route data
									pathInfo := &methodk8s.RouteInfo{
										Path: *match.Path.Value,
										Base: hostname,
										Service: &methodk8s.ServiceInfo{
											Name: service.Name,
											Uid:  string(service.UID),
											Namespace: &methodk8s.NamespaceInfo{
												Name: serviceNamespace.Name,
												Uid:  string(serviceNamespace.GetUID()),
											},
										},
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
				// HttpRoute data
				httpRouteInfo := &methodk8s.HttpRoute{
					Uid:         string(route.UID),
					Name:        route.GetName(),
					Namespace:   &namespaceInfo,
					Labels:      route.GetLabels(),
					Annotations: route.GetAnnotations(),
					Gateways:    gateways,
					Paths:       paths,
				}
				httpRoutes = append(httpRoutes, httpRouteInfo)
			}
		}
	}

	// Ingresses
	ingresses := []*methodk8s.Ingress{}
	if contains(types, "ingress") || len(types) == 0 {
		ingressList, err := clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			// Loop through Ingresses
			for _, ingress := range ingressList.Items {
				// Ingress Namespace
				namespace, err := clientset.CoreV1().Namespaces().Get(ctx, ingress.GetNamespace(), metav1.GetOptions{})
				if err != nil {
					errors = append(errors, err.Error())
					continue
				}
				namespaceInfo := methodk8s.NamespaceInfo{
					Name: namespace.GetName(),
					Uid:  string(namespace.GetUID()),
				}
				// Ingress routes
				rules := []*methodk8s.RouteInfo{}
				for _, rule := range ingress.Spec.Rules {
					basePath := rule.Host
					for _, path := range rule.HTTP.Paths {
						// Service
						service, err := clientset.CoreV1().Services(ingress.GetNamespace()).Get(ctx, rule.HTTP.Paths[0].Backend.Service.Name, metav1.GetOptions{})
						if err != nil {
							errors = append(errors, err.Error())
							continue
						}
						// Route data
						ruleInfo := methodk8s.RouteInfo{
							Path: path.Path,
							Base: basePath,
							Service: &methodk8s.ServiceInfo{
								Name: service.Name,
								Uid:  string(service.UID),
								Namespace: &methodk8s.NamespaceInfo{
									Name: namespace.GetName(),
									Uid:  string(namespace.GetUID()),
								},
							},
						}
						if service.Spec.Ports != nil && len(service.Spec.Ports) > 0 {
							port := fmt.Sprintf("%d", service.Spec.Ports[0].Port)
							ruleInfo.Port = &port
						}
						rules = append(rules, &ruleInfo)
					}
				}
				// Ingress 'catch all' route
				if ingress.Spec.DefaultBackend != nil {
					// Service
					service, err := clientset.CoreV1().Services(ingress.GetNamespace()).Get(ctx, ingress.Spec.DefaultBackend.Service.Name, metav1.GetOptions{})
					if err != nil {
						errors = append(errors, err.Error())
					} else {
						baseHost := ""
						if len(ingress.Spec.TLS) > 0 && len(ingress.Spec.TLS[0].Hosts) > 0 {
							baseHost = ingress.Spec.TLS[0].Hosts[0]
						}
						// Route data
						defaultRule := &methodk8s.RouteInfo{
							Path: "/",
							Base: baseHost,
							Service: &methodk8s.ServiceInfo{
								Name: service.Name,
								Uid:  string(service.UID),
								Namespace: &methodk8s.NamespaceInfo{
									Name: namespace.GetName(),
									Uid:  string(namespace.GetUID()),
								},
							},
						}
						if service.Spec.Ports != nil && len(service.Spec.Ports) > 0 {
							port := fmt.Sprintf("%d", service.Spec.Ports[0].Port)
							defaultRule.Port = &port
						}
						rules = append(rules, defaultRule)
					}
				}
				// Ingress Data
				ingressInfo := &methodk8s.Ingress{
					Uid:         string(ingress.UID),
					Name:        ingress.GetName(),
					Namespace:   &namespaceInfo,
					Rules:       rules,
					Annotations: ingress.GetAnnotations(),
					Labels:      ingress.GetLabels(),
				}
				ingresses = append(ingresses, ingressInfo)
			}
		}
	}

	// LoadBalancers
	loadBalancers := []*methodk8s.LoadBalancer{}
	if contains(types, "loadbalancer") || len(types) == 0 {
		// Fetch all services in all namespaces
		serviceList, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			// Loop through all Serivces that are 'LoadBalancers'
			serviceLoadBalancers := getLoadBalancerServices(serviceList.Items)
			for _, lb := range serviceLoadBalancers {
				namespace, err := clientset.CoreV1().Namespaces().Get(ctx, lb.GetNamespace(), metav1.GetOptions{})
				if err != nil {
					errors = append(errors, err.Error())
					continue
				}
				// LoadBalancer Namespace
				namespaceInfo := methodk8s.NamespaceInfo{
					Name: namespace.GetName(),
					Uid:  string(namespace.GetUID()),
				}
				// LoadBalancer routes
				paths := []*methodk8s.RouteInfo{}
				for _, ingress := range lb.Status.LoadBalancer.Ingress {
					basePath := ingress.IP
					if basePath == "" {
						basePath = ingress.Hostname
					}
					for _, port := range lb.Spec.Ports {
						portStr := fmt.Sprintf("%d", port.Port)
						// Route data
						pathInfo := methodk8s.RouteInfo{
							Path: "/",
							Base: basePath,
							Port: &portStr,
							Service: &methodk8s.ServiceInfo{
								Uid:       string(lb.UID),
								Name:      lb.Name,
								Namespace: &namespaceInfo,
							},
						}
						paths = append(paths, &pathInfo)
					}
				}
				// LoadBalancer Data
				loadBalancerInfo := &methodk8s.LoadBalancer{
					Uid:         string(lb.UID),
					Name:        lb.GetName(),
					Namespace:   &namespaceInfo,
					Annotations: lb.GetAnnotations(),
					Labels:      lb.GetLabels(),
					Paths:       paths,
				}
				loadBalancers = append(loadBalancers, loadBalancerInfo)

			}
		}
	}

	if len(httpRoutes) == 0 && len(ingresses) == 0 && len(loadBalancers) == 0 {
		return &methodk8s.IngressReport{AuthType: authType, Errors: errors}, nil
	}

	resources = methodk8s.IngressReport{
		HttpRoutes:    httpRoutes,
		Ingresses:     ingresses,
		LoadBalancers: loadBalancers,
		ClusterUrl:    &config.Host,
		AuthType:      authType,
		Errors:        errors,
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

func getLoadBalancerServices(serviceList []corev1.Service) []corev1.Service {
	var loadBalancerServices []corev1.Service
	for _, svc := range serviceList {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			loadBalancerServices = append(loadBalancerServices, svc)
		}
	}
	return loadBalancerServices
}
