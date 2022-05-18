package repo

import (
	"github.com/spf13/cobra"
	"skema-tool/internal/auth"
	"skema-tool/internal/pkg/console"
	"skema-tool/internal/pkg/repository"
)

const (
	shortGithubDescription = "manage github repo"
	longGithubDescription  = "sd repo github list"
)

func newGithubCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "github",
		Short: shortGithubDescription,
		Long:  longGithubDescription,
		Run: func(c *cobra.Command, args []string) {
			console.Info(longGithubDescription)
		},
	}

	cmd.AddCommand(newGithubListCmd())

	return cmd
}

func newGithubListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "list github repo",
		Long:  "list github repositories",
		Run: func(c *cobra.Command, args []string) {
			authProvider := auth.NewGithubAuthProvider()
			token := authProvider.GetLocalToken()
			repo := repository.NewGithubRepo(token)
			repoNames := repo.ListAvailableRepos()

			for i, n := range repoNames {
				console.Infof("%d: %s\n", i+1, n)
			}
		},
	}

	return cmd
}
