package project

import (
	"github.com/spf13/cobra"
	"skema-tool/internal/pkg/console"
)

const (
	projectDescription = "Generate Project Code"
)

func NewProjectCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "project",
		Short: "",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(projectDescription)
		},
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newInitCmd())

	return cmd
}
