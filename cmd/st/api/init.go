package api

import (
	"github.com/skema-dev/skema-tool/internal/api"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	initDescription     = "Init API Protocol Buffers Definition"
	initLongDescription = "sd api init --package=<package_name> --service=<service_name> --path=<output_path>"
)

func newInitCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init",
		Short: initDescription,
		Long:  initLongDescription,
		Run: func(c *cobra.Command, args []string) {
			userPackage, _ := c.Flags().GetString("package")
			userService, _ := c.Flags().GetString("service")
			path, _ := c.Flags().GetString("path")
			if path == "" {
				path = "./"
			}
			var userOptions []string
			optionValue, err := c.Flags().GetString("option")
			if err == nil {
				userOptions = strings.Split(optionValue, ";")
			}

			apiCreator := api.NewApiCreator()

			protoContent, err := apiCreator.InitProtoFile(
				userPackage,
				userService,
				userOptions,
			)
			protoFilepath := filepath.Join(path, userService+".proto")
			io.SaveToFile(protoFilepath, []byte(protoContent))
			console.Info(
				"New Protobuf file Created: %s\n==================\n%s\n",
				protoFilepath,
				protoContent,
			)
		},
	}

	cmd.Flags().StringP("package", "p", "", "package name")
	cmd.Flags().StringP("service", "s", "", "service name")
	cmd.Flags().String("path", "", "proto file path")
	cmd.Flags().String("option", "", "option for protobuf")

	cmd.MarkFlagRequired("package")
	cmd.MarkFlagRequired("service")

	return cmd
}
