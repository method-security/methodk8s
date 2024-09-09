package serviceaccount

import (
	"context"
	"encoding/base64"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintCredentials(ctx context.Context, k8sconfig *rest.Config, namespace string, secretname string) error {
	// Create the Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		return err
	}

	// Extract the Token from the secret
	secret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, secretname, metav1.GetOptions{})
	if err != nil {
		return err
	}

	token := base64.StdEncoding.EncodeToString(secret.Data["token"])
	caCert := base64.StdEncoding.EncodeToString(secret.Data["ca.crt"])
	apiServer := k8sconfig.Host

	// Pretty print the results
	fmt.Println("=================================")
	fmt.Println("Kubernetes Configuration Details")
	fmt.Println("=================================")
	fmt.Println("")
	fmt.Println("API Server URL:")
	fmt.Println("---------------------------------")
	fmt.Printf("%s\n\n", apiServer)
	fmt.Println("Token:")
	fmt.Println("---------------------------------")
	fmt.Printf("%s\n\n", token)
	fmt.Println("CA Certificate:")
	fmt.Println("---------------------------------")
	fmt.Printf("%s\n\n", caCert)
	fmt.Println("=================================")

	return nil
}
