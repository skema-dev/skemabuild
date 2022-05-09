package proto

import (
	"github.com/spf13/cobra"
	"skema-tool/internal/pkg/console"
)

const (
	publishDescription = "Publish Proto&Stub to Github"
)

func newPublishCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "publish",
		Short: "",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(publishDescription)
		},
	}

	return cmd
}
