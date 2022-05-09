package proto

import (
	"github.com/spf13/cobra"
	"skema-tool/internal/pkg/console"
)

const (
	description = "initialize protobuf file"
)

func NewProtoCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "proto",
		Short: "",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(description)
		},
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newPublishCmd())

	return cmd
}
