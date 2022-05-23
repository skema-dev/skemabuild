package service

import (
	"github.com/skema-dev/skema-tool/internal/pkg/console"

	"github.com/spf13/cobra"
)

const (
	createDescription     = "Create service code from api definition"
	createLongDescription = "sd service create --type=<framework> --tpl=<template> --api=<api_url>"
)

func newCreateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: createDescription,
		Long:  createLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			console.Info(createDescription)
		},
	}

	return cmd
}
