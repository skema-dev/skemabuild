package api

import (
	"io/ioutil"
	"path/filepath"
	"skema-tool/internal/pkg/io"
)

type StubCreator interface {
	GenerateStub(protobufContent string, packageOption string, outputPath string) (map[string]string, error)
}

func getCommonProtos() []string {
	protoNames := make([]string, 0)

	homePath := io.GetHomePath()
	commonProtoPath := filepath.Join(homePath, "common_protos")
	includeDirs, err := ioutil.ReadDir(commonProtoPath)
	if err != nil {
		panic(err)
	}

	for _, includeDir := range includeDirs {
		protoNames = append(protoNames, filepath.Join(commonProtoPath, includeDir.Name()))
	}
	return protoNames
}
