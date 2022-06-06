package api

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/skema-dev/skema-tool/internal/pkg/io"
)

func NewOpenapiStubCreator() StubCreator {
	return &openapiStubCreator{}
}

type openapiStubCreator struct {
}

// Verify if {{ .Options }} exists. If not, replace with package name
func (s *openapiStubCreator) tryAddGoPackageOption(content string) string {
	if strings.Contains(content, ProtocobufOptionTplStr) {
		// needs to replace option with go_package
		packageName := GetPackageNameFromProto(content)
		packageOption := fmt.Sprintf("option go_package=\"%s\";\n", packageName)
		newContent := strings.Replace(content, ProtocobufOptionTplStr, packageOption, 1)
		return newContent
	}

	return content
}

func (s *openapiStubCreator) Generate(protobufContent string) (map[string]string, error) {
	content := protobufContent
	content = s.tryAddGoPackageOption(content)

	// Create temporary path for stub files
	homePath := io.GetHomePath()
	tempPath := filepath.Join(homePath, "temp", "stub-gen", uuid.New().String(), "openapi")

	// genereate protoc arguments
	opts := make([]string, 0)
	opts = append(opts, fmt.Sprintf("-I=%s", tempPath),
		"--openapiv2_opt=use_go_templates=true",
		fmt.Sprintf("--openapiv2_out=%s", tempPath),
	)

	stubs, err := GenerateStub(content, tempPath, opts, true)
	return stubs, err
}
