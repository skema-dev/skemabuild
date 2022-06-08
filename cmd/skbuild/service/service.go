package service

import (
	"github.com/skema-dev/skemabuild/internal/pkg/console"

	"github.com/spf13/cobra"
)

const (
	serviceDescription     = "Generate Service Code"
	serviceLongDescription = "skbuild service create --type=grpc-go --tpl=skema-grpc --api=github.com/xxxxx/test.pb"
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
