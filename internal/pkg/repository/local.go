package repository

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type localRepo struct {
	repo         *git.Repository
	localPath    string
	relativePath string
}

func NewLocalRepo(localRepoPath string, relativePath string) Repository {
	repo, err := git.PlainOpen(localRepoPath)
	if err != nil {
		console.Fatalf(err.Error())
	}

	return &localRepo{
		repo:         repo,
		localPath:    localRepoPath,
		relativePath: relativePath,
	}
}

func checkIfError(err error) {
	if err != nil {
		console.Fatalf(err.Error())
	}
}

func (r *localRepo) UploadToRepo(files map[string]string, repoPath string, forceCreateNewRepo bool) (string, error) {
	// write files to current path (supposing this
	w, err := r.repo.Worktree()
	checkIfError(err)

	console.Info("Files to be commited")
	commitFiles := make(map[string]string)
	for f, v := range files {
		newPath := filepath.Join(repoPath, f)
		commitFiles[newPath] = v
		console.Info(newPath)
		io.SaveToFile(newPath, []byte(v))
		_, err = w.Add(newPath)
		checkIfError(err)
	}

	_, err = w.Commit("upload stubs", &git.CommitOptions{})
	checkIfError(err)
	console.Info("start push...")

	publicKey := r.publicKey()
	err = r.repo.Push(&git.PushOptions{
		Auth: publicKey,
	})
	checkIfError(err)

	return "", nil
}

func (r *localRepo) AddVersion(repoName string, version string, commitID string) error {
	if ok, err := r.setTag(version); !ok {
		console.Fatalf(err.Error())
	}
	err := r.pushTags()
	checkIfError(err)
	return nil
}

func (r *localRepo) ListAvailableRepos() []string {
	return nil
}

func (r *localRepo) GetContents(repoName, path string, opts ...string) (result map[string]string, err error) {
	return nil, nil
}

func (r *localRepo) checkExistingTag(tag string) {
	tags, err := r.repo.TagObjects()
	checkIfError(err)

	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name == tag {
			console.Fatalf("tag %s already exists", t.Name)
		}
		return nil
	})
}

func (r *localRepo) setTag(tag string) (bool, error) {
	r.checkExistingTag(tag)

	console.Info("Set tag %s", tag)
	h, err := r.repo.Head()
	checkIfError(err)

	_, err = r.repo.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Message: tag,
	})
	checkIfError(err)

	return true, nil
}

func (r *localRepo) pushTags() error {
	publicKey := r.publicKey()

	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       publicKey,
	}
	err := r.repo.Push(po)
	checkIfError(err)
	return nil
}

func (r *localRepo) publicKey() *ssh.PublicKeys {
	sshPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	sshKey, _ := ioutil.ReadFile(sshPath)
	publicKey, err := ssh.NewPublicKeys("git", []byte(sshKey), "")
	checkIfError(err)

	return publicKey
}
