package repository

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/skema-dev/skemabuild/internal/pkg/io"
)

type localRepo struct {
	repo         *git.Repository
	localPath    string
	relativePath string
	username     string
	password     string
}

func NewLocalRepo(
	localRepoPath string,
	relativePath string,
	username string,
	password string,
) Repository {
	repo, err := git.PlainOpen(localRepoPath)
	if err != nil {
		console.Fatalf(err.Error())
	}

	return &localRepo{
		repo:         repo,
		localPath:    localRepoPath,
		relativePath: relativePath,
		username:     username,
		password:     password,
	}
}

func (r *localRepo) UploadToRepo(
	files map[string]string,
	repoPath string,
	forceCreateNewRepo bool,
) (string, error) {
	w, err := r.repo.Worktree()
	console.FatalIfError(err)

	console.Info("Files to be commited")
	commitFiles := make(map[string]string)
	for f, v := range files {
		newPath := filepath.Join(repoPath, f)
		commitFiles[newPath] = v
		console.Info(newPath)
		io.SaveToFile(newPath, []byte(v))
		_, err = w.Add(newPath)
		console.FatalIfError(err)
	}

	_, err = w.Commit("upload stubs", &git.CommitOptions{})
	console.FatalIfError(err)
	console.Info("start push...")

	err = r.repo.Push(&git.PushOptions{
		Auth: r.authMethod(),
	})
	console.FatalIfError(err)

	return "", nil
}

func (r *localRepo) AddVersion(repoName string, version string, commitID string) error {
	if ok, err := r.setTag(version); !ok {
		console.Fatalf(err.Error())
	}
	err := r.pushTags()
	console.FatalIfError(err)
	return nil
}

func (r *localRepo) ListAvailableRepos() []string {
	return nil
}

func (r *localRepo) GetContents(
	repoName, path string,
	opts ...string,
) (result map[string]string, err error) {
	return nil, nil
}

func (r *localRepo) checkExistingTag(tag string) {
	tags, err := r.repo.TagObjects()
	console.FatalIfError(err)

	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name == tag {
			console.Fatalf("tag %s already exists", t.Name)
		}
		return nil
	})
}

func (r *localRepo) setTag(tag string) (bool, error) {
	r.checkExistingTag(tag)

	h, err := r.repo.Head()
	console.FatalIfError(err)

	_, err = r.repo.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Message: tag,
	})
	console.FatalIfError(err)

	return true, nil
}

func (r *localRepo) pushTags() error {
	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       r.authMethod(),
	}
	err := r.repo.Push(po)
	console.FatalIfError(err)
	return nil
}

func (r *localRepo) authMethod() transport.AuthMethod {
	if r.username != "" && r.password != "" {
		auth := &http.BasicAuth{
			Username: r.username,
			Password: r.password,
		}
		console.Info("use http auth for git")
		return auth
	}

	sshPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	sshKey, _ := ioutil.ReadFile(sshPath)
	publicKey, err := ssh.NewPublicKeys("git", []byte(sshKey), "")
	console.FatalIfError(err)
	console.Info("use ssh auth for git")

	return publicKey
}
