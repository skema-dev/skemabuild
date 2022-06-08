package repo

import (
	"github.com/spf13/cobra"
)

const (
	shortDescription = "display repository information"
	longDescription  = "skbuild repo github list"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "repo",
		Short: shortDescription,
		Long:  longDescription,
		Run: func(c *cobra.Command, args []string) {
		},
	}

	cmd.AddCommand(newGithubCmd())

	return cmd
}
