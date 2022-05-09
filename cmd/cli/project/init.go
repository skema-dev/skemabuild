package project

import (
	"github.com/spf13/cobra"
	"skema-tool/internal/pkg/console"
)

const (
	initDescription = "Initialize project layout"
)

func newInitCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init",
		Short: "",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(initDescription)
		},
	}

	return cmd
}
