package dev

import (
	"os"
	"path/filepath"

	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/spf13/cobra"
)

const (
	clusterDescription       = "skbuild dev cluster -h"
	clusterCreateDescription = "skbuild dev cluster init [--name <cluster-name>] [-f <kind-cluster-init.yaml>]"
	clusterDeleteDescription = "skbuild dev cluster destroy [--name <cluster-name>]"

	defaultInitScript = "cluster_init.sh"
	defaultScriptPath = "./env/dev"
)

func newClusterCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cluster",
		Short: clusterDescription,
		Long:  clusterDescription,
		Run: func(c *cobra.Command, args []string) {
			console.Info(clusterDescription)
		},
	}

	cmd.AddCommand(newClusterCreateCmd())
	cmd.AddCommand(newClusterDeleteCmd())

	return cmd
}

func newClusterCreateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: clusterCreateDescription,
		Long:  clusterCreateDescription,
		Run: func(c *cobra.Command, args []string) {
			clusterName := c.Flag("name").Value.String()
			configFile := c.Flag("file").Value.String()

			scriptPath := filepath.Join(defaultScriptPath, defaultInitScript)
			if _, err := os.Stat(scriptPath); err != nil {
				console.Fatalf("Cannot find %s. Please execute the command in the root directory of the project.", scriptPath)
			}

			arguments := []string{defaultInitScript}
			if clusterName != "" {
				arguments = append(arguments, clusterName)
			}
			if configFile != "" {
				arguments = append(arguments, configFile)
			}
			err := console.ExecCommandInPath(defaultScriptPath, "sh", arguments...)
			console.FatalIfError(err)
		},
	}

	cmd.Flags().StringP("name", "n", "", "name of the cluster")
	cmd.Flags().StringP("file", "f", "kind-cluster.yaml", "path to the cluster init file")

	return cmd
}

func newClusterDeleteCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete",
		Short: clusterDeleteDescription,
		Long:  clusterDeleteDescription,
		Run: func(c *cobra.Command, args []string) {
			clusterName := c.Flag("name").Value.String()
			arguments := []string{"delete", "cluster"}
			if clusterName != "" {
				arguments = append(arguments, "--name", clusterName)

			}
			err := console.ExecCommand("kind", arguments...)
			if err != nil {
				console.Errorf("%s\n", err.Error())
			}
		},
	}

	cmd.Flags().StringP("name", "n", "", "name of the cluster")

	return cmd
}
