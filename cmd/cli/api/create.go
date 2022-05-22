package api

import (
	"github.com/skema-dev/skema-go/logging"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"skema-tool/internal/api"
	"skema-tool/internal/pkg/console"
	"skema-tool/internal/pkg/io"
	"strings"
)

const (
	createDescription     = "Create API Stubs"
	createLongDescription = "st api create --go_option github.com/com/test --input ./Hello1.proto -o ./stub-test"
)

func newCreateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: createDescription,
		Long:  createLongDescription,
		Run: func(c *cobra.Command, args []string) {
			input, _ := c.Flags().GetString("input")
			output, _ := c.Flags().GetString("output")
			stubTypes, _ := c.Flags().GetString("type")
			goOption, _ := c.Flags().GetString("go_option")

			if debug, _ := c.Flags().GetBool("debug"); debug {
				logging.Init("debug", "console")
			}

			//TODO: download remote file if input starts with http[s]://

			data, _ := os.ReadFile(input)
			content := string(data)

			stubTypeArr := strings.Split(stubTypes, ",")
			for _, stubType := range stubTypeArr {
				stubType = strings.TrimRight(strings.TrimLeft(stubType, " "), " ")
				outputPath := filepath.Join(output, stubType)
				var creator api.StubCreator

				switch stubType {
				case "grpc-go":
					creator = api.NewGoStubCreator(goOption)
				case "openapi":
					creator = api.NewOpenapiStubCreator()
				default:
					console.Errorf("unsupported stub type: %s", stubType)
					continue
				}

				console.Infof("[%s]\n", stubType)
				stubs, err := creator.Generate(content)
				if err != nil {
					panic(err.Error())
				}

				for filename, stub := range stubs {
					stubFilepath := filepath.Join(outputPath, filename)
					if err := io.SaveToFile(stubFilepath, []byte(stub)); err != nil {
						panic(err)
					}
					console.Infof("%s\n", stubFilepath)
				}
			}
		},
	}

	cmd.Flags().StringP("input", "i", "", "path of input protobuf file")
	cmd.Flags().StringP("output", "o", "./", "output path for generated stubs")
	cmd.Flags().String("go_option", "", "go_package option to be used in stub")
	cmd.Flags().StringP("type", "t", "grpc-go,openapi", "stub types to generate.")
	cmd.Flags().Bool("debug", false, "enable debug output")

	cmd.MarkFlagRequired("input")

	return cmd
}
