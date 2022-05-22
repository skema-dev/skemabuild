package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v44/github"
	"github.com/google/uuid"
	"github.com/skema-dev/skema-go/logging"
	"golang.org/x/oauth2"
	"skema-tool/internal/pkg/console"
	"skema-tool/internal/pkg/pattern"
	"strings"
	"time"
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

func ParseGithubUrl(url string) (string, string, string) {
	r := "(https://){0,1}github\\.com/(?P<organization_name>[a-zA-Z0-9-_]+)/(?P<repo_name>[a-zA-Z0-9-_]+)/(tree/main/){0,1}(?P<repo_path>[a-zA-Z0-9.\\-_\\/]+)"
	found := pattern.GetNamedMapFromText(url, r, []string{"organization_name", "repo_name", "repo_path"})

	return found["organization_name"], found["repo_name"], found["repo_path"]
}

// Most of the supporting functions are from https://github.com/google/go-github/blob/master/example/commitpr/main.go
func (g *GithubRepo) UploadToRepo(files map[string]string, repoUrl string, forceCreateNewRepo bool) (string, error) {
	console.Info("start parsing url")
	organization, repoName, repoPath := ParseGithubUrl(repoUrl)
	if organization == "" || repoName == "" {
		return "", errors.New("incorrect github organization or repo name")
	}

	console.Info("org: %s\nrepo:%s\n", organization, repoName)

	currentTime := time.Now()
	commitMessage := fmt.Sprintf("Upload Stub at %s", currentTime.Format("2017-09-07-17-06-06"))
	commitBranch := strings.ReplaceAll(uuid.New().String(), "-", "")

	repo, err := g.getRepository(repoName, organization, forceCreateNewRepo)
	if err != nil {
		return "", err
	}

	baseBranch := repo.GetDefaultBranch()
	if baseBranch == "" {
		panic("cannot find default branch for repo " + repoName)
	}

	prTitle := commitMessage

	if len(files) == 0 {
		return "", errors.New("no files included to upload")
	}

	ref, err := g.getRef(repoName, commitBranch, baseBranch)
	if err != nil {
		return "", err
	}
	if ref == nil {
		return "", fmt.Errorf("git ref is nil. something is wrong for commit branch %s", commitBranch)
	}

	tree, err := g.getTreeToCommit(repoName, ref, repoPath, files)
	if err != nil {
		return "", err
	}

	if err = g.pushCommit(repoName, ref, tree, commitMessage); err != nil {
		return "", err
	}

	pr, err := g.createPR(repoName, prTitle, "", commitBranch, baseBranch)
	if err != nil {
		return "", err
	}

	pr, err = g.mergePR(repoName, pr, commitMessage)
	if err != nil {
		return "", err
	}

	if err = g.removeBranch(repoName, commitBranch); err != nil {
		console.Errorf("Error when removing branch %s: %s\n", commitBranch, err.Error())
	}

	return pr.GetMergeCommitSHA(), nil
}

func (g *GithubRepo) AddVersion(repoName string, version string, commitID string) error {
	_, err := g.createRef(repoName, "refs/tags/"+version, commitID)
	return err
}

func (g *GithubRepo) ListAvailableRepos() []string {
	repos := make([]string, 0)

	repositories, _, err := g.client.Repositories.List(g.ctx, g.username, nil)
	if err != nil {
		console.Errorf("list available github repos error: %s\n", err.Error())
		return repos
	}

	for _, r := range repositories {
		repos = append(repos, *r.FullName)
	}

	return repos
}

func (g *GithubRepo) getRepository(repoName string, organization string, forceNewRepo bool) (*github.Repository, error) {
	logging.Debugf("prepare github repo %s/%s", organization, repoName)
	repo, _, err := g.client.Repositories.Get(g.ctx, g.username, repoName)
	if err == nil {
		return repo, nil
	}

	if !forceNewRepo {
		return nil, fmt.Errorf("repo %s doesn't exist", repoName)
	}

	if organization != g.username {
		return nil, fmt.Errorf("creating new repo in orgnization should be done via github")
	}

	owner, _, err := g.client.Users.Get(g.ctx, g.username)

	// create new repo
	console.Info("creating new repo: %s", repoName)
	repo = &github.Repository{Owner: owner, Name: github.String(repoName)}
	repo, _, err = g.client.Repositories.Create(g.ctx, "", repo)
	if err != nil {
		return nil, err
	}

	// init repo by adding readme
	console.Info("add readme to init")
	file := "README.md"
	createFileOpts := &github.RepositoryContentFileOptions{
		Content: []byte("This repo is initialized by skema"),
		Message: github.String("# TODO"),
	}
	rsp, _, err := g.client.Repositories.CreateFile(g.ctx, g.username, repoName, file, createFileOpts)
	if err != nil {
		return nil, err
	}

	// commit the file
	commit := &github.Commit{
		Message: github.String("Initial commit"),
		Tree:    rsp.Tree,
	}
	commit, _, err = g.client.Git.CreateCommit(g.ctx, g.username, repoName, commit)
	if err != nil {
		return nil, err
	}

	// update ref
	ref := &github.Reference{
		Object: &github.GitObject{
			SHA: commit.SHA,
		},
		Ref: github.String("heads/" + *repo.DefaultBranch),
	}
	_, _, err = g.client.Git.UpdateRef(g.ctx, g.username, repoName, ref, true)
	if err != nil {
		return nil, err
	}

	return repo, err
}

func (g *GithubRepo) getRef(repoName string, commitBranch string, baseBranch string) (ref *github.Reference, err error) {
	if ref, _, err = g.client.Git.GetRef(g.ctx, g.username, repoName, "refs/heads/"+commitBranch); err == nil {
		return ref, nil
	}

	if commitBranch == baseBranch {
		return nil, errors.New(
			"the commit branch does not exist but `-base-branch` is the same as `-commit-branch`",
		)
	}

	var baseRef *github.Reference
	if baseRef, _, err = g.client.Git.GetRef(g.ctx, g.username, repoName, "refs/heads/"+baseBranch); err != nil {
		return nil, err
	}
	return g.createRef(repoName, "refs/heads/"+commitBranch, *baseRef.Object.SHA)
}

func (g *GithubRepo) createRef(repoName string, ref string, sha string) (*github.Reference, error) {
	newRef := &github.Reference{
		Ref:    github.String(ref),
		Object: &github.GitObject{SHA: github.String(sha)}}
	reference, _, err := g.client.Git.CreateRef(g.ctx, g.username, repoName, newRef)
	return reference, err
}

// getTree generates the tree to commit based on the given files and the commit
// of the ref you got in getRef.
func (g *GithubRepo) getTreeToCommit(repoName string, ref *github.Reference, rootPath string, files map[string]string) (tree *github.Tree, err error) {
	logging.Debugf("prepare files to commit from %s", rootPath)
	// Create a tree with what to commit.
	entries := []*github.TreeEntry{}

	// Load each file into the tree.
	for relativePath, content := range files {
		path := fmt.Sprintf("%s/%s", rootPath, relativePath)
		entries = append(entries, &github.TreeEntry{
			Path:    github.String(path),
			Type:    github.String("blob"),
			Content: github.String(content),
			Mode:    github.String("100644")})
	}

	tree, _, err = g.client.Git.CreateTree(g.ctx, g.username, repoName, *ref.Object.SHA, entries)
	return tree, err
}

// pushCommit creates the commit in the given reference using the given tree.
func (g *GithubRepo) pushCommit(
	repoName string,
	ref *github.Reference,
	tree *github.Tree,
	commitMessage string) (err error) {
	logging.Debugf("push commit to github")
	// Get the parent commit to attach the commit to.
	parent, _, err := g.client.Repositories.GetCommit(g.ctx, g.username, repoName, *ref.Object.SHA, nil)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	user, _, err := g.client.Users.Get(g.ctx, "")
	if err != nil {
		return err
	}

	var email string
	if user.Email != nil {
		email = *user.Email
	} else {
		email, err = g.getUserEmail()
		if err != nil {
			return err
		}
		if email == "" {
			email = "dev@skema.dev"
		}
	}

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: user.Login, Email: github.String(email)}
	commit := &github.Commit{
		Author:  author,
		Message: &commitMessage,
		Tree:    tree,
		Parents: []*github.Commit{parent.Commit},
	}
	newCommit, _, err := g.client.Git.CreateCommit(g.ctx, g.username, repoName, commit)
	if err != nil {
		return err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = g.client.Git.UpdateRef(g.ctx, g.username, repoName, ref, false)
	return err
}

// createPR creates a pull request. Based on:
// https://godoc.org/github.com/google/go-github/github#example-PullRequestsService-Create
func (g *GithubRepo) createPR(
	repoName string,
	prSubject string,
	prDescription string,
	commitBranch string,
	targetBranch string) (p *github.PullRequest, err error) {
	logging.Debugf("create pull request on temp branch %s", commitBranch)
	newPR := &github.NewPullRequest{
		Title:               &prSubject,
		Head:                &commitBranch,
		Base:                &targetBranch,
		Body:                &prDescription,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := g.client.PullRequests.Create(g.ctx, g.username, repoName, newPR)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (g *GithubRepo) getUserEmail() (string, error) {
	userEmails, _, err := g.client.Users.ListEmails(g.ctx, &github.ListOptions{
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		return "", err
	}
	if len(userEmails) == 0 {
		return "", nil
	}
	return *userEmails[0].Email, nil
}

func (g *GithubRepo) mergePR(
	repoName string,
	pr *github.PullRequest,
	mergeMessage string) (*github.PullRequest, error) {
	logging.Debugf("merge pull request")
	pullRequestNumber := *pr.Number
	opts := &github.PullRequestOptions{
		MergeMethod: "squash",
	}
	result, _, err := g.client.PullRequests.Merge(
		g.ctx,
		g.username,
		repoName,
		pullRequestNumber,
		mergeMessage,
		opts,
	)
	if err != nil {
		return nil, err
	}
	if !result.GetMerged() {
		return nil, errors.New("Failed to merge")
	}

	pr, _, err = g.client.PullRequests.Get(g.ctx, g.username, repoName, pullRequestNumber)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (g *GithubRepo) removeBranch(repoName, branchName string) error {
	logging.Debugf("remove temp branch %s", branchName)
	_, err := g.client.Git.DeleteRef(g.ctx, g.username, repoName, "refs/heads/"+branchName)
	return err
}
