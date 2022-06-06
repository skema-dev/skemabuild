package service

import (
	"github.com/skema-dev/skema-tool/internal/auth"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/repository"
	"github.com/skema-dev/skema-tool/internal/service"
	"github.com/spf13/cobra"
)

const (
	createDescription     = "Create service code from api definition"
	createLongDescription = "sd service create --type=<framework> --tpl=<template> --protobuf=<protobuf_uri>"
)

func newCreateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: createDescription,
		Long:  createLongDescription,
		Run: func(c *cobra.Command, args []string) {
			protoFile, err := c.Flags().GetString("protobuf")
			if err != nil {
				console.Fatalf("invalid protobuf definition")
			}
			console.Info("protobuf: %s", protoFile)

			authProvider := auth.NewGithubAuthProvider()
			repo := repository.NewGithubRepo(authProvider.GetLocalToken())
			if repo == nil {
				console.Fatalf("failed to initiate github repo")
			}
			repoName, repoPath := service.GetRemoteProtobufLocation(protoFile)
			console.Info("Repo: %s\nPath: %s", repoName, repoPath)

			content, err := repo.GetContents(repoName, repoPath)
			if err != nil {
				console.Fatalf(err.Error())
			}
			service.GetProtobufDescriptionFromString(content[repoPath])

		},
	}

	cmd.Flags().StringP("protobuf", "p", "", "protobuf file")
	cmd.MarkFlagRequired("protobuf")

	return cmd
}
