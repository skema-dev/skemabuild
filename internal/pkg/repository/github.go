package repository

import (
	"context"
	"github.com/google/go-github/v42/github"
	"golang.org/x/oauth2"
	"skema-tool/internal/pkg/console"
)

type GithubRepo struct {
	token    string
	username string
	client   *github.Client
	ctx      context.Context
}

func NewGithubRepo(token string) Repository {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		console.Errorf("failed to get github user:%s\n", err.Error())
		return nil
	}
	username := user.Login

	repo := &GithubRepo{
		token:    token,
		client:   client,
		ctx:      ctx,
		username: *username,
	}
	return repo
}

func (g *GithubRepo) UploadToRepo(files map[string]string, repoUrl string, repoPath string) error {
	return nil
}

func (g *GithubRepo) ListAvailableRepos() []string {
	repos := make([]string, 0)

	repositories, _, err := g.client.Repositories.ListAll(g.ctx, nil)
	if err != nil {
		console.Errorf("list available github repos error: %s\n", err.Error())
		return repos
	}

	for _, r := range repositories {
		repos = append(repos, *r.FullName)
	}

	return repos
}
