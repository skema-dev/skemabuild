package service

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/skema-dev/skemabuild/internal/auth"
	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/skema-dev/skemabuild/internal/pkg/http"
	"github.com/skema-dev/skemabuild/internal/pkg/io"
	"github.com/skema-dev/skemabuild/internal/pkg/pattern"
	"github.com/skema-dev/skemabuild/internal/pkg/repository"
	"github.com/skema-dev/skemabuild/internal/service"
	"github.com/spf13/cobra"
)

const (
	createDescription     = "Create service code from protocol buffers definition"
	createLongDescription = "skbuild service create --proto=<protobuf_uri>"
)

func newCreateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: createDescription,
		Long:  createLongDescription,
		Run: func(c *cobra.Command, args []string) {
			protoUrl := c.Flag("proto").Value.String()
			goModule := c.Flag("module").Value.String()
			goVersion, _ := c.Flags().GetString("goversion")
			serviceName, _ := c.Flags().GetString("service")
			output, _ := c.Flags().GetString("output")
			tpl, _ := c.Flags().GetString("tpl")
			s, _ := c.Flags().GetString("http")
			httpEnabled, _ := strconv.ParseBool(s)
			userParams, _ := c.Flags().GetString("parameters")

			userParameters := map[string]string{}
			if userParams != "" {
				ss := strings.Split(userParams, ",")
				for _, s := range ss {
					kv := strings.Split(s, ":")
					if len(kv) != 2 {
						console.Fatalf("Invalid parameter: %s", s)
					}
					k := kv[0]
					v := kv[1]
					userParameters[k] = v
				}
			}

			var rpcParameters *service.RpcParameters
			if pattern.IsGithubUrl(protoUrl) {
				// use github client to get proto file
				authProvider := auth.NewGithubAuthProvider()
				repo := repository.NewGithubRepo(authProvider.GetLocalToken())
				if repo == nil {
					console.Fatalf("failed to initiate github repo")
				}
				repoName, repoPath, _ := service.GetGithubContentLocation(protoUrl)
				console.Info("get remote proto on github: %s", protoUrl)
				console.Info("Repo: %s\nPath: %s", repoName, repoPath)

				content, err := repo.GetContents(repoName, repoPath)
				if err != nil {
					console.Fatalf(err.Error())
				}
				rpcParameters = service.GetRpcParameters(
					content[repoPath],
					goModule,
					goVersion,
					serviceName,
				)
			} else if pattern.IsHttpUrl(protoUrl) {
				// get proto by regular http
				console.Info("get remote proto: %s", protoUrl)
				content := http.GetTextContent(protoUrl)
				rpcParameters = service.GetRpcParameters(content, goModule, goVersion, serviceName)
			} else {
				// read from local path
				data, err := os.ReadFile(protoUrl)
				console.FatalIfError(err)
				content := string(data)
				rpcParameters = service.GetRpcParameters(content, goModule, goVersion, serviceName)
			}
			rpcParameters.HttpEnabled = httpEnabled

			generator := service.NewGrpcGoGenerator()
			contents := generator.CreateCodeContent(tpl, rpcParameters, userParameters)

			for path, c := range contents {
				outputPath := filepath.Join(output, path)
				io.SaveToFile(outputPath, []byte(c))
				console.Info(outputPath)
			}
		},
	}

	cmd.Flags().StringP("proto", "p", "", "protobuf file")
	cmd.Flags().StringP("module", "m", "", "go module name")
	cmd.Flags().StringP("goversion", "v", "1.16", "go version")
	cmd.Flags().StringP("service", "s", "", "service name")
	cmd.Flags().StringP("tpl", "t", "standard", "template name or url")
	cmd.Flags().String("http", "true", "enable http or not")
	cmd.Flags().StringP("output", "o", "", "output path")
	cmd.Flags().String("parameter", "", "user defined tpl parameters: key1:value1,key2:value2...")
	cmd.MarkFlagRequired("proto")

	return cmd
}
