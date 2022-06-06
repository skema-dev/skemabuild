package repository

type Repository interface {
	UploadToRepo(files map[string]string, repoUrl string, forceCreateNewRepo bool) (string, error)
	AddVersion(repoName string, version string, commitID string) error
	ListAvailableRepos() []string
	GetContents(repoName, path string, opts ...string) (result map[string]string, err error)
}
