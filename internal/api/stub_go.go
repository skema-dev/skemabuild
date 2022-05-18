package api

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"path/filepath"
	"skema-tool/internal/pkg/console"
	"skema-tool/internal/pkg/io"
	"strings"
)

func NewGoStubCreator() StubCreator {
	return &goStubCreator{}
}

type goStubCreator struct {
}

func (s *goStubCreator) Generate(protobufContent string, packageOption string) (map[string]string, error) {
	if packageOption == "" {
		return nil, errors.New("must define go_package=xxx")
	}

	// Create temporary path for stub files
	homePath := io.GetHomePath()
	tempPath := filepath.Join(homePath, "temp", "stub-gen", uuid.New().String(), "go")

	// add go_package in protobuf
	newPackageOption := fmt.Sprintf("option go_package=\"%s\";\n", packageOption)
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
