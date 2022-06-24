package dev

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/spf13/cobra"
)

const (
	serviceDescription           = "skbuild dev service -h"
	serviceImageBuildDescription = "skbuild dev service build --name <service_name>"
	serviceCreateDescription     = "skbuild dev service create --name <service_name>"
	serviceDeleteDescription     = "skbuild dev service delete --name <service_name>"

	defaultConfigRootPath = "./env/dev"
)

func newServiceCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "service",
		Short: serviceDescription,
		Long:  serviceDescription,
		Run: func(c *cobra.Command, args []string) {
			console.Info(serviceDescription)
		},
	}

	cmd.AddCommand(newServiceImageBuildCmd())
	cmd.AddCommand(newServiceCreateCmd())
	cmd.AddCommand(newServiceDeleteCmd())

	return cmd
}

func newServiceImageBuildCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "build",
		Short: serviceImageBuildDescription,
		Long:  serviceImageBuildDescription,
		Run: func(c *cobra.Command, args []string) {
			serviceName := c.Flag("name").Value.String()

			dockerfilePath := filepath.Join(defaultConfigRootPath, serviceName, "Dockerfile")
			if _, err := os.Stat(dockerfilePath); err != nil {
				console.Fatalf("Cannot find %s. Please execute the command in the root directory of the project.", dockerfilePath)
			}

			arguments := []string{"build", "-t", fmt.Sprintf("dev/%s", serviceName), "-f", dockerfilePath, "."}
			err := console.ExecCommand("docker", arguments...)
			console.FatalIfError(err)
		},
	}

	cmd.Flags().StringP("name", "n", "", "name of the service")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newServiceCreateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: serviceCreateDescription,
		Long:  serviceCreateDescription,
		Run: func(c *cobra.Command, args []string) {
			serviceName := c.Flag("name").Value.String()

			scriptPath := filepath.Join(defaultConfigRootPath, serviceName, "deploy.sh")
			if _, err := os.Stat(scriptPath); err != nil {
				console.Fatalf("Cannot find %s. Please execute the command in the root directory of the project.", scriptPath)
			}

			err := console.ExecCommandInPath(filepath.Join(defaultConfigRootPath, serviceName), "sh", "deploy.sh")
			console.FatalIfError(err)

		},
	}

	cmd.Flags().StringP("name", "n", "", "name of the cluster")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newServiceDeleteCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete",
		Short: serviceDeleteDescription,
		Long:  serviceDeleteDescription,
		Run: func(c *cobra.Command, args []string) {
			serviceName := c.Flag("name").Value.String()
			createOrDeleteFromDeployConfig(serviceName, "delete")
		},
	}

	cmd.Flags().StringP("name", "n", "", "name of the service")
	cmd.MarkFlagRequired("name")

	return cmd
}

func createOrDeleteFromDeployConfig(serviceName string, operation string) {
	configPath := filepath.Join(defaultConfigRootPath, serviceName, "deploy.yaml")
	if _, err := os.Stat(configPath); err != nil {
		console.Fatalf("Cannot find %s. Please execute the command in the root directory of the project.", configPath)
	}

	arguments := []string{operation, "-f", configPath}
	err := console.ExecCommand("kubectl", arguments...)
	console.FatalIfError(err)
}
