package generator

import (
	"github.com/skema-dev/skemabuild/internal/pkg/pattern"
)

func GetGithubContentLocation(url string) (string, string, string) {
	r := "https://github\\.com/(?P<organization_name>[a-zA-Z0-9-_]+)/(?P<repo_name>[a-zA-Z0-9-_]+)/(blob/main/){0,1}(tree/main/){0,1}(?P<repo_path>[a-zA-Z0-9-_/.]+)"
	resourceMap := pattern.GetNamedMapFromText(
		url,
		r,
		[]string{"organization_name", "repo_name", "repo_path"},
	)
	repoName := resourceMap["repo_name"]
	repoPath := resourceMap["repo_path"]
	organization := resourceMap["organization_name"]
	return repoName, repoPath, organization
}
