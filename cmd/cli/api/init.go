package api

import (
	"path/filepath"
	"skema-tool/internal/api"
	"skema-tool/internal/pkg/console"
	"skema-tool/internal/pkg/io"
	"strings"

	"github.com/spf13/cobra"
)

const (
	initDescription     = "Init API Protocol Buffers Definition"
	initLongDescription = "sd api init --module=<module_name> --package=<package_name> --service=<service_name> --path=<output_path> --option=<protobuf options>"
)

func newInitCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init",
		Short: initDescription,
		Long:  initLongDescription,
		Run: func(c *cobra.Command, args []string) {
			userModule, _ := c.Flags().GetString("module")
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
				userModule,
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

	cmd.Flags().StringP("module", "m", "", "module name")
	cmd.Flags().StringP("package", "p", "", "package name")
	cmd.Flags().StringP("service", "s", "", "service name")
	cmd.Flags().String("path", "", "proto file path")
	cmd.Flags().String("option", "", "option for protobuf")

	cmd.MarkFlagRequired("module")
	cmd.MarkFlagRequired("package")
	cmd.MarkFlagRequired("service")

	return cmd
}
