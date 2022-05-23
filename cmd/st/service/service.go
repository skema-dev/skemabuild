package service

import (
	"github.com/skema-dev/skema-tool/internal/pkg/console"

	"github.com/spf13/cobra"
)

const (
	serviceDescription     = "Generate Service Code"
	serviceLongDescription = "sd service create --type=grpc-go --tpl=skema-grpc --api=github.com/xxxxx/test.pb"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "service",
		Short: serviceDescription,
		Long:  serviceLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(serviceDescription)
		},
	}

	cmd.AddCommand(newCreateCmd())

	return cmd
}
