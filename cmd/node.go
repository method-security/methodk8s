package cmd

import (
	"github.com/method-security/methodk8s/internal/node"
	"github.com/spf13/cobra"
)

func (a *MethodK8s) InitNodeCommand() {
	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Audit and command Nodes",
		Long:  `Audit and command Nodes`,
	}

	enumerateCmd := &cobra.Command{
		Use:   "enumerate",
		Short: "Enumerate Nodes",
		Long:  `Enumerate Nodes`,
		Run: func(cmd *cobra.Command, args []string) {
			// Init Cmd Failed
			if a.OutputSignal.ErrorMessage != nil {
				return
			}

			report, err := node.EnumerateNodes(cmd.Context(), a.K8sConfig, a.AuthType)
			if err != nil {
				errorMessage := err.Error()
				a.OutputSignal.ErrorMessage = &errorMessage
				a.OutputSignal.Status = 1
			}
			a.OutputSignal.Content = report
		},
	}

	nodeCmd.AddCommand(enumerateCmd)
	a.RootCmd.AddCommand(nodeCmd)
}
