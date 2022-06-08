package main

import (
	"fmt"
	"github.com/skema-dev/skema-go/logging"
	"github.com/skema-dev/skemabuild/cmd/skbuild/api"
	"github.com/skema-dev/skemabuild/cmd/skbuild/auth"
	"github.com/skema-dev/skemabuild/cmd/skbuild/repo"
	"github.com/skema-dev/skemabuild/cmd/skbuild/service"

	"github.com/spf13/cobra"
)

const (
	description = "SkemaBuild(skbuild): Build Protocol Buffers Based gRPC Service Code"
	version     = "preview"
)

func main() {
	rootCmd := newCmdRoot()
	_ = rootCmd.Execute()
}

func newCmdRoot() *cobra.Command {
	logging.Init("info", "console")
	var cmd = &cobra.Command{
		Use:   "skbuild",
		Short: fmt.Sprintf("SkemaBuild(skbuild) version(%s)", version),
		Long:  fmt.Sprintf("\n%s\nversion: %s", description, version),
	}

	cmd.AddCommand(auth.NewCmd())
	cmd.AddCommand(api.NewCmd())
	cmd.AddCommand(service.NewCmd())
	cmd.AddCommand(repo.NewCmd())

	return cmd
}
