package api

import (
	"bytes"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/io"
	"github.com/skema-dev/skema-tool/internal/pkg/repository"
)

//go:embed tpl/go_mod.tpl
var goModTemplate string

func NewGoStubCreator(packageOption string) StubCreator {
	if packageOption == "" {
		console.Fatalf("must define go package option")
	}

	return &goStubCreator{
		packageOption: packageOption,
	}
}

type goStubCreator struct {
	packageOption string
}

func (s *goStubCreator) Generate(protobufContent string) (map[string]string, error) {
	// Create temporary path for stub files
	homePath := io.GetHomePath()
	tempPath := filepath.Join(homePath, "temp", "stub-gen", uuid.New().String(), "go")

	// add go_package in protobuf
	newPackageOption := fmt.Sprintf("option go_package=\"%s\";\n", s.packageOption)
	content := strings.Replace(protobufContent, ProtocobufOptionTplStr, newPackageOption, 1)

	// genereate protoc arguments
	opts := make([]string, 0)
	opts = append(opts, fmt.Sprintf("-I=%s", tempPath),
		"--go_opt=paths=source_relative",
		"--go-grpc_opt=paths=source_relative",
		"--validate_opt=paths=source_relative",
		"--grpc-gateway_opt=paths=source_relative",
		"--grpc-gateway_opt=generate_unbound_methods=true",
		fmt.Sprintf("--go_out=%s", tempPath),
		fmt.Sprintf("--go-grpc_out=%s", tempPath),
		fmt.Sprintf("--validate_out=lang=go:%s", tempPath),
		fmt.Sprintf("--grpc-gateway_out=%s", tempPath),
	)

	result, err := GenerateStub(content, tempPath, opts, true)
	if err != nil {
		console.Errorf("Go Stub Generating Failed: %s\n", err.Error())
		return nil, err
	}

	return result, nil
}

func GetExpectedGithubGoPackageUri(uploadUrl string, protobufContent string) string {
	organization, repoName, repoPath := repository.ParseGithubUrl(uploadUrl)
	if organization == "" || repoName == "" {
		console.Fatalf("incorrect github url definition")
	}
	packageName := GetPackageNameFromProto(protobufContent)
	packagePath := fmt.Sprintf(
		"github.com/%s/%s/%s/%s/grpc-go",
		organization,
		repoName,
		repoPath,
		packageName,
	)
	return packagePath
}

func GenerateGoMod(packagePath string) string {
	tpl := template.Must(template.New("default").Option("missingkey=zero").Parse(goModTemplate))

	userValues := map[string]string{
		"PackageAddress": strings.ToLower(packagePath),
		"GoVersion":      "1.16",
	}

	var content bytes.Buffer
	err := tpl.Execute(&content, userValues)
	if err != nil {
		return ""
	}

	return content.String()
}
