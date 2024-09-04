package cmd

import (
	"github.com/method-security/methodk8s/internal/serviceaccount"
	"github.com/spf13/cobra"
)

func (a *MethodK8s) InitServiceAccountCommand() {
	serviceAccountCmd := &cobra.Command{

		Use:   "serviceaccount",
		Short: "Configure, audit and command Service Accounts",
		Long:  `Configure, audit and command Service Accounts`,
	}

	configureAccountCmd := &cobra.Command{

		Use:   "configure",
		Short: "Configure Service Account",
		Long:  `Configure Service Account`,
	}

	credsCmd := &cobra.Command{
		Use:   "creds",
		Short: "Service account credentials",
		Long:  `Use this command to print the Service Account credentials`,
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
				return
			}

			secretname, err := cmd.Flags().GetString("secretname")
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
				return
			}

			err = serviceaccount.PrintCredentials(cmd.Context(), a.K8sConfig, namespace, secretname)
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
			}
			a.OutputSignal.Content = nil
		},
	}
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Create a service account in your k8s cluster",
		Long:  `Create a service account in your k8s cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			run, err := cmd.Flags().GetBool("run")
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
				return
			}

			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
				return
			}

			err = serviceaccount.Config(cmd.Context(), a.K8sConfig, run, namespace)
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
			}
			a.OutputSignal.Content = nil
		},
	}

	credsCmd.Flags().String("secretname", "method-sa-secret", "The name of the secret to use for authentication (Defaults to method-sa-secret)")
	credsCmd.Flags().String("namespace", "default", "Set the namespace for the Service Account and Secret (Defaults to 'default')")
	applyCmd.Flags().Bool("run", false, "Apply the Service Account yamls (Defaults to False)")
	applyCmd.Flags().String("namespace", "default", "Set the namespace for the Service Account and Secret (Defaults to 'default')")

	configureAccountCmd.AddCommand(credsCmd)
	configureAccountCmd.AddCommand(applyCmd)
	serviceAccountCmd.AddCommand(configureAccountCmd)
	a.RootCmd.AddCommand(serviceAccountCmd)
}
