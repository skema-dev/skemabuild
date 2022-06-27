package api

import (
	"path/filepath"
	"strings"

	"github.com/skema-dev/skemabuild/internal/api"
	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/skema-dev/skemabuild/internal/pkg/io"

	"github.com/spf13/cobra"
)

const (
	initDescription     = "Init API Protocol Buffers Definition"
	initLongDescription = "skbuild api init --package=<package_name> --service=<service_name> --path=<output_path>"
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

			if !strings.Contains(userPackage, ".") && !strings.Contains(userPackage, "/") {
				console.Fatalf("Invalide package name. package must contains . or /, not a single word")
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
				"New Protobuf file Created: %s\n==================\n%s\n==================\n%s\n",
				protoFilepath,
				protoContent,
				"change the {{ .Options }} to your package address if necessary.",
			)

			setupFilepath := filepath.Join(path, userService+".yaml")
			setupYamlContent, err := apiCreator.InitSetupFile(userService)
			io.SaveToFile(setupFilepath, []byte(setupYamlContent))
			console.Infof(setupFilepath + " created. \nModify the file and then execute `skbuild -f service.yaml` to generate everything\n")
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
