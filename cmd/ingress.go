package cmd

import (
	"github.com/method-security/methodk8s/internal/ingress"
	"github.com/spf13/cobra"
)

func (a *MethodK8s) InitIngressCommand() {
	ingressCmd := &cobra.Command{
		Use:   "ingress",
		Short: "Audit and command Ingresses",
		Long:  `Audit and command Ingresses`,
	}

	enumerateCmd := &cobra.Command{
		Use:   "enumerate",
		Short: "Enumerate Ingresses",
		Long:  `Enumerate Ingresses`,
		Run: func(cmd *cobra.Command, args []string) {

			types, err := cmd.Flags().GetStringSlice("type")
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
				return
			}

			report, err := ingress.EnumerateIngresses(cmd.Context(), a.K8sConfig, a.AuthType, types)
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
			}
			a.OutputSignal.Content = report
		},
	}
	enumerateCmd.Flags().StringSlice("type", []string{}, "List the types to emumerate (ie.--type ingress --type gateway --type loadbalancer)")

	ingressCmd.AddCommand(enumerateCmd)
	a.RootCmd.AddCommand(ingressCmd)
}
