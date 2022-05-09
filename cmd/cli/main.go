package main

import (
	"github.com/spf13/cobra"
	"skema-tool/cmd/cli/project"
	"skema-tool/cmd/cli/proto"
)

const (
	description = "toolkit for quick protobuf based service code generating"
)

func main() {
	rootCmd := newCmdRoot()
	_ = rootCmd.Execute()
}

func newCmdRoot() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sd",
		Short: description,
		Long:  description,
	}

	cmd.AddCommand(proto.NewProtoCmd())
	cmd.AddCommand(project.NewProjectCmd())

	return cmd
}
