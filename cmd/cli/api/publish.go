package api

import (
	"skema-tool/internal/pkg/console"

	"github.com/spf13/cobra"
)

const (
	publishDescription     = "Publish Proto&Stub to Github"
	publishLongDescription = "sd api publish --stubpath=<stub filepath> --type=<repo_type> --group=<user|orgnization> --repo=<repo_name> --path=<path_in_repo>"
)

func newPublishCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "publish",
		Short: publishDescription,
		Long:  publishLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(publishDescription)
		},
	}

	return cmd
}
