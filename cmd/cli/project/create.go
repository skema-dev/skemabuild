package project

import (
	"github.com/spf13/cobra"
	"skema-tool/internal/pkg/console"
)

const (
	createDescription = "Create service code from protobuf file"
)

func newCreateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: "",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(createDescription)
		},
	}

	return cmd
}
