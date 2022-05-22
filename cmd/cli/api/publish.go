package api

import (
	"fmt"
	"github.com/skema-dev/skema-go/logging"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"skema-tool/internal/api"
	"skema-tool/internal/auth"
	"skema-tool/internal/pkg/console"
	"skema-tool/internal/pkg/repository"
	"strings"
)

const (
	publishDescription     = "Publish Proto&Stub to Github"
	publishLongDescription = "st api publish --input=./stub-test --url https://github.com/likezhang-public/newst/tes1 --version=v0.0.1"
)

func newPublishCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "publish",
		Short: publishDescription,
		Long:  publishLongDescription,
		Run: func(c *cobra.Command, args []string) {
			input, _ := c.Flags().GetString("input")
			uploadUrl, _ := c.Flags().GetString("url")
			version, _ := c.Flags().GetString("version")
			if debug, _ := c.Flags().GetBool("debug"); debug {
				logging.Init("debug", "console")
			}

			authProvider := auth.NewGithubAuthProvider()
			repo := repository.NewGithubRepo(authProvider.GetLocalToken())
			if repo == nil {
				console.Fatalf("not able to get autenticated repo. please double check.")
			}
			goPackage := ""
			originalPackageName := ""

			stubs := make(map[string]string)
			// iterate temp path, and return all file contents
			// IMPORTANT: generate and insert go.mod for go package
			err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
				// read file path
				if info.IsDir() {
					return nil
				}

				relativePath, err := filepath.Rel(input, path)
				if err != nil {
					panic(fmt.Sprintf("incorrect input path %s: %s", input, err.Error()))
				}

				data, err := ioutil.ReadFile(path)
				if err != nil {
					logging.Errorf("failed to read %s\n", path)
					return err
				}
				stubs[relativePath] = string(data)

				if goPackage == "" {
					if strings.HasPrefix(relativePath, "grpc-go") && strings.HasSuffix(relativePath, ".proto") {
						// make sure go package definition is compatible with upload url
						content := stubs[relativePath]
						goPackage = api.GetOptionGoPackageNameFromProto(content)
						expectedPackage := api.GetExpectedGithubGoPackageName(uploadUrl, content)
						if goPackage != expectedPackage {
							console.Fatalf("Incorrect package definition\nCurrent go_package=\"%s\"\nExpected go_package=\"%s\"\n", goPackage, expectedPackage)
						}
						goModContent := api.GenerateGoMod(goPackage)
						stubs["grpc-go/go.mod"] = goModContent
						originalPackageName = api.GetPackageNameFromProto(content)
					}
				}

				return nil
			})
			if err != nil {
				panic(err)
			}

			// attach original package name after given path in github repo
			uploadUrl = fmt.Sprintf("%s/%s", uploadUrl, originalPackageName)
			console.Info("uploading to %s", uploadUrl)
			_, repoName, repoPath := repository.ParseGithubUrl(uploadUrl)
			commitID, err := repo.UploadToRepo(stubs, uploadUrl, true)
			if err != nil {
				console.Fatalf(err.Error())
			}

			// add tag for the published version
			githubVersionTag := fmt.Sprintf("%s/grpc-go/%s", repoPath, version)
			console.Info("update version to %s", githubVersionTag)
			if err := repo.AddVersion(repoName, githubVersionTag, commitID); err != nil {
				console.Fatalf("failed to create new version: %s", err.Error())
			}

			// output the new version to be imported in go project
			console.Info("new version published: go get %s@%s", goPackage, version)
		},
	}
	cmd.Flags().StringP("input", "i", "", "path of input stubs")
	cmd.Flags().StringP("url", "o", "", "github url to upload")
	cmd.Flags().String("version", "", "version to be published")
	cmd.Flags().Bool("debug", false, "enable debug output")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("url")
	cmd.MarkFlagRequired("version")

	return cmd
}
