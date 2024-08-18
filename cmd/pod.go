package cmd

import (
	"github.com/method-security/methodk8s/internal/pod"
	"github.com/spf13/cobra"
)

func (a *MethodK8s) InitPodCommand() {
	podCmd := &cobra.Command{
		Use:   "pod",
		Short: "Audit and command Pods",
		Long:  `Audit and command Pods`,
	}

	enumerateCmd := &cobra.Command{
		Use:   "enumerate",
		Short: "Enumerate Pods",
		Long:  `Enumerate Pods`,
		Run: func(cmd *cobra.Command, args []string) {
			report, err := pod.EnumeratePods(cmd.Context(), a.K8sConfig, a.AuthType)
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
			}
			a.OutputSignal.Content = report
		},
	}

	podCmd.AddCommand(enumerateCmd)
	a.RootCmd.AddCommand(podCmd)
}
