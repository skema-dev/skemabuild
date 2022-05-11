package api

import (
	"skema-tool/internal/pkg/console"

	"github.com/spf13/cobra"
)

const (
	createDescription     = "Create API Stubs"
	createLongDescription = "sd api create --input=<protobuf_file_path> --output=<output_path>"
)

func newCreateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: createDescription,
		Long:  createLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(createDescription)
		},
	}

	return cmd
}
