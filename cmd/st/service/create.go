package service

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/skema-dev/skema-tool/internal/auth"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/io"
	"github.com/skema-dev/skema-tool/internal/pkg/repository"
	"github.com/skema-dev/skema-tool/internal/service"
	"github.com/spf13/cobra"
)

const (
	createDescription     = "Create service code from api definition"
	createLongDescription = "sd service create --type=<framework> --tpl=<template> --proto=<protobuf_uri>"
)

func newCreateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: createDescription,
		Long:  createLongDescription,
		Run: func(c *cobra.Command, args []string) {
			protoUrl, err := c.Flags().GetString("proto")
			if err != nil {
				console.Fatalf("invalid protobuf definition")
			}
			goModule, _ := c.Flags().GetString("module")
			goVersion, _ := c.Flags().GetString("goversion")
			serviceName, _ := c.Flags().GetString("service")
			output, _ := c.Flags().GetString("output")
			tpl, _ := c.Flags().GetString("tpl")
			s, _ := c.Flags().GetString("http")
			httpEnabled, _ := strconv.ParseBool(s)

			userParameters := map[string]string{}

			var rpcParameters *service.RpcParameters
			if strings.HasPrefix(protoUrl, "https://github.com/") {
				// use github client to get proto file
				authProvider := auth.NewGithubAuthProvider()
				repo := repository.NewGithubRepo(authProvider.GetLocalToken())
				if repo == nil {
					console.Fatalf("failed to initiate github repo")
				}
				repoName, repoPath, _ := service.GetGithubContentLocation(protoUrl)
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
			} else {
				// get proto by regular http
				client := resty.New()
				resp, _ := client.R().
					Get(protoUrl)
				content := string(resp.Body())
				rpcParameters = service.GetRpcParameters(content, goModule, goVersion, serviceName)
			}
			rpcParameters.HttpEnabled = httpEnabled

			console.Infof("http: %t\n", httpEnabled)

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
	cmd.Flags().StringP("tpl", "t", "simple", "template name or url")
	cmd.Flags().String("http", "true", "enable http or not")
	cmd.Flags().StringP("output", "o", "", "output path")
	cmd.MarkFlagRequired("proto")

	return cmd
}
