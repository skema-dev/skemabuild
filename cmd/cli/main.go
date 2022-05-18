package main

import (
	"skema-tool/cmd/cli/api"
	"skema-tool/cmd/cli/auth"
	"skema-tool/cmd/cli/repo"
	"skema-tool/cmd/cli/service"

	"github.com/spf13/cobra"
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

	cmd.AddCommand(auth.NewCmd())
	cmd.AddCommand(api.NewCmd())
	cmd.AddCommand(service.NewCmd())
	cmd.AddCommand(repo.NewCmd())

	return cmd
}
