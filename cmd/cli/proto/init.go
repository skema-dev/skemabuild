package proto

import (
	"github.com/spf13/cobra"
	"path/filepath"
	"skema-tool/internal/pkg/console"
	"skema-tool/internal/pkg/io"
	"skema-tool/internal/proto"
	"strings"
)

const (
	initDescription = "Init protobuf file"
)

func newInitCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init",
		Short: "",
		Long:  "",
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

			protoTool := proto.Proto{}

			protoContent, err := protoTool.InitProtoFile(userModule, userPackage, userService, userOptions)
			protoFilepath := filepath.Join(path, userService+".proto")
			io.SaveToFile(protoFilepath, []byte(protoContent))
			console.Info("New Protofile Created: %s\n==================\n%s\n", protoFilepath, protoContent)
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
