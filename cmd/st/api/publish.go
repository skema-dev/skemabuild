package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/skema-dev/skema-go/logging"
	"github.com/skema-dev/skema-tool/internal/api"
	"github.com/skema-dev/skema-tool/internal/auth"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/repository"
	"github.com/spf13/cobra"
)

const (
	publishDescription     = "Publish Proto&Stub to Github"
	publishLongDescription = "st api publish --stub=./stub-test --url https://github.com/likezhang-public/newst/tes1 --version=v0.0.1"
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

			if stub != "" && proto != "" {
				console.Fatalf("cannot specify both stub and proto!")
			}

			if stub != "" {
				publishFromStubs(stub, uploadUrl, version)
			} else if proto != "" {
				publishFromProto(proto, stubTypes, uploadUrl, version)
			}
		},
	}
	cmd.Flags().StringP("stub", "s", "", "path of input stubs")
	cmd.Flags().StringP("proto", "p", "", "path of input stubs")
	cmd.Flags().StringP("url", "u", "", "github url to upload")
	cmd.Flags().String("version", "v", "version to be published")
	cmd.Flags().StringP("type", "t", "grpc-go,openapi", "stub types to generate.")
	cmd.Flags().Bool("debug", false, "enable debug output")

	cmd.MarkFlagRequired("url")
	cmd.MarkFlagRequired("version")

	return cmd
}

func publishFromStubs(stubPath string, uploadUrl string, version string) {
	stubs, goPackage, originalPackageName, err := loadUploadingStubs(stubPath, uploadUrl)
	if err != nil {
		console.Fatalf("Not able to load local stubs from %s. %s", stubPath, err.Error())
	}

	// attach original package name after given path in github repo
	stubUploadUrl := fmt.Sprintf("%s/%s", uploadUrl, originalPackageName)
	uploadGithubStubsAndTagVersion(stubUploadUrl, stubs, version)

	// output the new version to be imported in go project
	console.Info("new version published: go get %s@%s", goPackage, version)
}

func publishFromProto(protoPath string, stubTypes string, uploadUrl string, version string) {
	data, err := os.ReadFile(protoPath)
	if err != nil {
		console.Fatalf(err.Error())
	}
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
				// make sure go package definition is compatible with upload url
				content := stubs[relativePath]
				goPackage = api.GetOptionGoPackageNameFromProto(content)
				expectedPackage := api.GetExpectedGithubGoPackageUri(uploadUrl, content)
				if goPackage != expectedPackage {
					console.Fatalf(
						"Incorrect package definition\nCurrent go_package=\"%s\"\nExpected go_package=\"%s\"\n",
						goPackage,
						expectedPackage,
					)
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
