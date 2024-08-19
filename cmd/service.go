package cmd

import (
	"github.com/method-security/methodk8s/internal/service"
	"github.com/spf13/cobra"
)

func (a *MethodK8s) InitServiceCommand() {
	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "Audit and command Services",
		Long:  `Audit and command Services`,
	}

	enumerateCmd := &cobra.Command{
		Use:   "enumerate",
		Short: "Enumerate Services",
		Long:  `Enumerate Services`,
		Run: func(cmd *cobra.Command, args []string) {
			report, err := service.EnumerateServices(cmd.Context(), a.K8sConfig, a.AuthType)
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
			}
			a.OutputSignal.Content = report
		},
	}

	serviceCmd.AddCommand(enumerateCmd)
	a.RootCmd.AddCommand(serviceCmd)
}
