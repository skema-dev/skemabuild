package api

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"path/filepath"
	"skema-tool/internal/pkg/console"
	"skema-tool/internal/pkg/io"
	"strings"
)

const (
	ProtocoBufOptionTplStr = `{{ .Options }}`
)

func NewGoStubCreator() StubCreator {
	return &goStubCreator{}
}

type goStubCreator struct {
}

func (s *goStubCreator) GenerateStub(protobufContent string, packageOption string, outputPath string) (map[string]string, error) {
	if packageOption == "" {
		return nil, errors.New("must define go_package=xxx")
	}

	// Create temporary path for stub files
	homePath := io.GetHomePath()
	tempPath := filepath.Join(homePath, "temp", "stub-gen", uuid.New().String())

	// add go_package in protobuf
	protoFilePath := filepath.Join(tempPath, "input.proto")
	content := strings.Replace(protobufContent, ProtocoBufOptionTplStr, "option go_package=\""+packageOption+"\";\n", 1)
	io.SaveToFile(protoFilePath, []byte(content))

	// genereate protoc arguments
	opts := make([]string, 0)
	commonProtos := getCommonProtos()
	for _, opt := range commonProtos {
		opts = append(opts, fmt.Sprintf("-I=%s", opt))
	}
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
		protoFilePath,
	)

	// execute protoc
	console.Info("exec cmd: %s",
		fmt.Sprintf("protoc %s", strings.Join(opts, " ")))
	err := console.ExecCommand("protoc", opts...)
	if err != nil {
		console.Errorf(err.Error())
		return nil, err
	}

	// iterate all stubs and save into map. Using stub filename relative to the temp path
	// as the key to return
	stubs := make(map[string]string)
	// iterate temp path, and return all file contents
	err = filepath.Walk(tempPath, func(path string, info os.FileInfo, err error) error {
		// read file path
		if info.IsDir() {
			return nil
		}
		relativePath := strings.TrimPrefix(path, tempPath)[1:]
		console.Infof("generated stub: %s\n", relativePath)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			console.Errorf("failed to read %s\n", path)
			return err
		}
		stubs[relativePath] = string(data)
		return nil
	})

	os.RemoveAll(tempPath)
	return stubs, err
}
