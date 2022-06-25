package main

import (
	"fmt"
	"github.com/skema-dev/skema-go/logging"
	"github.com/skema-dev/skemabuild/cmd/skbuild/api"
	"github.com/skema-dev/skemabuild/cmd/skbuild/auth"
	"github.com/skema-dev/skemabuild/cmd/skbuild/dev"
	"github.com/skema-dev/skemabuild/cmd/skbuild/repo"
	"github.com/skema-dev/skemabuild/cmd/skbuild/service"
	"github.com/skema-dev/skemabuild/internal/pkg/console"

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
		Run: func(c *cobra.Command, args []string) {
			configFile, _ := c.Flags().GetString("file")
			if configFile == "" {
				console.Info(c.Long)
				return
			}
			username, _ := c.Flags().GetString("username")
			password, _ := c.Flags().GetString("password")

			buildServiceFromConfigFile(configFile, username, password)
		},
	}

	cmd.AddCommand(auth.NewCmd())
	cmd.AddCommand(api.NewCmd())
	cmd.AddCommand(service.NewCmd())
	cmd.AddCommand(repo.NewCmd())
	cmd.AddCommand(dev.NewCmd())

	cmd.Flags().StringP("file", "f", "", "config file to generate service code")
	cmd.Flags().String("username", "", "git username for http auth")
	cmd.Flags().String("password", "", "git password for http auth")

	return cmd
}
