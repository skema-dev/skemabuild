package api

import (
	"github.com/skema-dev/skema-tool/internal/pkg/console"

	"github.com/spf13/cobra"
)

const (
	description = "manage api definitions"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "api",
		Short: description,
		Long:  description,
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(description)
		},
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newPublishCmd())

	return cmd
}
