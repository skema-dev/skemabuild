package repository

type Repository interface {
	UploadToRepo(files map[string]string, repoUrl string, repoPath string) error
	ListAvailableRepos() []string
}
