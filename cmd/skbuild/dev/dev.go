package dev

import (
	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/spf13/cobra"
)

const (
	shortDescription = "dev environment setup"
	longDescription  = "provide shortcuts to manage local kubernetes cluster"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dev",
		Short: shortDescription,
		Long:  longDescription,
		Run: func(c *cobra.Command, args []string) {
			console.Info(longDescription)
		},
	}

	cmd.AddCommand(newClusterCmd())
	cmd.AddCommand(newServiceCmd())

	return cmd
}
