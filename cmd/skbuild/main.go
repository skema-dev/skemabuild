package main

import (
	"github.com/skema-dev/skema-go/logging"
	"github.com/skema-dev/skemabuild/cmd/skbuild/api"
	"github.com/skema-dev/skemabuild/cmd/skbuild/auth"
	"github.com/skema-dev/skemabuild/cmd/skbuild/repo"
	"github.com/skema-dev/skemabuild/cmd/skbuild/service"

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
	logging.Init("info", "console")
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
