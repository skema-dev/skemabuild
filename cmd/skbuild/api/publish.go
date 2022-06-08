package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/skema-dev/skema-go/logging"
	"github.com/skema-dev/skemabuild/internal/api"
	"github.com/skema-dev/skemabuild/internal/auth"
	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/skema-dev/skemabuild/internal/pkg/repository"
	"github.com/spf13/cobra"
)

const (
	publishDescription     = "Publish Proto&Stub to Git"
	publishLongDescription = `
1. publish to github
skbuild api publish --stub=./stub-test --url https://github.com/likezhang-public/newst/tes1 --version=v0.0.1
2. publish to relative path in current repo

skbuild api publish --stub=./stub-test --url <path_in_repo> --version=v0.0.1

3. publish using http auth
skbuild api publish --stub=./stub-test --url <path_in_repo> --version=v0.0.1 --username=<username> --password=<password>
`
)

func newPublishCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "publish",
		Short: publishDescription,
		Long:  publishLongDescription,
		Run: func(c *cobra.Command, args []string) {
			stub, _ := c.Flags().GetString("stub")
			proto, _ := c.Flags().GetString("proto")
			uploadUrl, _ := c.Flags().GetString("url")
			version, _ := c.Flags().GetString("version")
			if debug, _ := c.Flags().GetBool("debug"); debug {
				logging.Init("debug", "console")
			}
			stubTypes, _ := c.Flags().GetString("type")
			username, _ := c.Flags().GetString("username")
			password, _ := c.Flags().GetString("password")

			if stub != "" && proto != "" {
				console.Fatalf("cannot specify both stub and proto!")
			}

			if stub != "" {
				publishFromStubs(stub, uploadUrl, version, username, password)
			} else if proto != "" {
				publishFromProto(proto, stubTypes, uploadUrl, version, username, password)
			}
		},
	}
	cmd.Flags().StringP("stub", "s", "", "path of input stubs")
	cmd.Flags().StringP("proto", "p", "", "path of input stubs")
	cmd.Flags().StringP("url", "u", "", "github url to upload")
	cmd.Flags().String("version", "v", "version to be published")
	cmd.Flags().StringP("type", "t", "grpc-go,openapi", "stub types to generate.")
	cmd.Flags().String("username", "", "git username for http auth")
	cmd.Flags().String("password", "", "git password for http auth")
	cmd.Flags().Bool("debug", false, "enable debug output")

	cmd.MarkFlagRequired("url")
	cmd.MarkFlagRequired("version")

	return cmd
}

func publishFromStubs(
	stubPath string,
	uploadUrl string,
	version string,
	username string,
	password string,
) {
	stubs, goPackage, originalPackageName, err := loadUploadingStubs(stubPath, uploadUrl)
	if err != nil {
		console.Fatalf("Not able to load local stubs from %s. %s", stubPath, err.Error())
	}

	switch getRepoTypeFromUrl(uploadUrl) {
	case "github":
		// attach original package name after given path in github repo
		stubUploadUrl := fmt.Sprintf("%s/%s", uploadUrl, originalPackageName)
		uploadGithubStubsAndTagVersion(stubUploadUrl, stubs, version)
	default:
		stubRepoPath := fmt.Sprintf("%s/%s", uploadUrl, originalPackageName)
		uploadLocalStubsAndTagVersion(stubRepoPath, stubs, version, username, password)
	}

	// output the new version to be imported in go project
	console.Info("new version published: go get %s@%s", goPackage, version)
}

func publishFromProto(
	protoPath string,
	stubTypes string,
	uploadUrl string,
	version string,
	username string,
	password string,
) {
	data, err := os.ReadFile(protoPath)
	if err != nil {
		console.Fatalf(err.Error())
	}
	switch getRepoTypeFromUrl(uploadUrl) {
	case "github":
		protoContent := string(data)
		expectedPackageUri := api.GetExpectedGithubGoPackageUri(uploadUrl, protoContent)
		stubs, err := generateStubsFromProto(protoPath, stubTypes, expectedPackageUri)
		if err != nil {
			console.Fatalf(err.Error())
		}

		// attach original package name after given path in github repo
		stubUploadUrl := fmt.Sprintf("%s/%s", uploadUrl, api.GetPackageNameFromProto(protoContent))
		uploadGithubStubsAndTagVersion(stubUploadUrl, stubs, version)

		// output the new version to be imported in go project
		console.Info("new version published: go get %s@%s", expectedPackageUri, version)
	default:
		protoContent := string(data)
		expectedPackageUri := api.GetOptionPackageNameFromProto(protoContent, "go_package")
		stubs, err := generateStubsFromProto(protoPath, stubTypes, expectedPackageUri)
		if err != nil {
			console.Fatalf(err.Error())
		}

		// attach original package name after given path in github repo
		stubRepoPath := fmt.Sprintf("%s/%s", uploadUrl, api.GetPackageNameFromProto(protoContent))
		uploadLocalStubsAndTagVersion(stubRepoPath, stubs, version, username, password)

		// output the new version to be imported in go project
		console.Info("new version published: go get %s@%s", expectedPackageUri, version)
	}
}

func loadUploadingStubs(
	inputPath string,
	uploadUrl string,
) (stubs map[string]string, goPackage string, originalPackageName string, err error) {
	// iterate temp path, and return all file contents
	// IMPORTANT: generate and insert go.mod for go package
	stubs = make(map[string]string)

	err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		// read file path
		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(inputPath, path)
		if err != nil {
			panic(fmt.Sprintf("incorrect input path %s: %s", inputPath, err.Error()))
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			logging.Errorf("failed to read %s\n", path)
			return err
		}
		stubs[relativePath] = string(data)

		if goPackage == "" {
			if strings.HasPrefix(relativePath, "grpc-go") &&
				strings.HasSuffix(relativePath, ".proto") {
				// make sure go package definition is not empty
				content := stubs[relativePath]
				goPackage = api.GetOptionGoPackageNameFromProto(content)
				if getRepoTypeFromUrl(uploadUrl) == "github" {
					// validate go_package only for github upload
					expectedPackage := api.GetExpectedGithubGoPackageUri(uploadUrl, content)
					if goPackage != expectedPackage {
						console.Fatalf(
							"Incorrect package definition\nCurrent go_package=\"%s\"\nExpected go_package=\"%s\"\n",
							goPackage,
							expectedPackage,
						)
					}
				}
				goModContent := api.GenerateGoMod(goPackage)
				stubs["grpc-go/go.mod"] = goModContent
				originalPackageName = api.GetPackageNameFromProto(content)
			}
		}

		return nil
	})

	return stubs, goPackage, originalPackageName, err
}

func getRepoTypeFromUrl(uploadUrl string) string {
	if strings.HasPrefix(uploadUrl, "http://github.com") ||
		strings.HasPrefix(uploadUrl, "https://github.com") {
		return "github"
	}

	return "local"
}

func uploadGithubStubsAndTagVersion(stubUploadUrl string, stubs map[string]string, version string) {
	authProvider := auth.NewGithubAuthProvider()
	repo := repository.NewGithubRepo(authProvider.GetLocalToken())
	if repo == nil {
		console.Fatalf("not able to get autenticated repo. please double check.")
	}

	console.Info("uploading to %s", stubUploadUrl)
	_, repoName, repoPath := repository.ParseGithubUrl(stubUploadUrl)
	commitID, err := repo.UploadToRepo(stubs, stubUploadUrl, true)
	if err != nil {
		console.Fatalf(err.Error())
	}

	// add tag for the published version
	githubVersionTag := fmt.Sprintf("%s/grpc-go/%s", repoPath, version)
	console.Info("update version to %s", githubVersionTag)
	if err := repo.AddVersion(repoName, githubVersionTag, commitID); err != nil {
		console.Fatalf("failed to create new version: %s", err.Error())
	}
}

func uploadLocalStubsAndTagVersion(
	stubRepoPath string,
	stubs map[string]string,
	version string,
	username string,
	password string,
) {
	currentPath, err := os.Getwd()
	if err != nil {
		console.Fatalf(err.Error())
	}

	repo := repository.NewLocalRepo(currentPath, stubRepoPath, username, password)
	if repo == nil {
		console.Fatalf("Make sure your current path is at the root of a git repo")
	}

	console.Info("uploading to %s", stubRepoPath)
	_, err = repo.UploadToRepo(stubs, stubRepoPath, false)
	if err != nil {
		console.Fatalf(err.Error())
	}

	// add tag for the published version
	githubVersionTag := fmt.Sprintf("%s/grpc-go/%s", stubRepoPath, version)
	console.Info("update version to %s", githubVersionTag)
	if err := repo.AddVersion("", githubVersionTag, ""); err != nil {
		console.Fatalf("failed to create new version: %s", err.Error())
	}
}
