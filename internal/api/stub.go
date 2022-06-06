package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/skema-dev/skema-go/logging"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/io"
	"github.com/skema-dev/skema-tool/internal/pkg/pattern"
)

const (
	ProtocobufOptionTplStr = `{{ .Options }}`
)

type StubCreator interface {
	Generate(protobufContent string) (map[string]string, error)
}

func getCommonProtos() []string {
	protoNames := make([]string, 0)

	homePath := io.GetHomePath()
	commonProtoPath := filepath.Join(homePath, "protos")
	includeDirs, err := ioutil.ReadDir(commonProtoPath)
	if err != nil {
		panic(err)
	}

	for _, includeDir := range includeDirs {
		protoNames = append(protoNames, filepath.Join(commonProtoPath, includeDir.Name()))
	}
	return protoNames
}

func GetOptionGoPackageNameFromProto(protoContent string) string {
	return GetOptionPackageNameFromProto(protoContent, "go_package")
}

func GetOptionPackageNameFromProto(protoContent string, packageName string) string {
	reg := fmt.Sprintf(
		"option[\\s]+%s=\"(?P<option_package_name>[a-zA-Z0-9.\\-_\\/]+)\";",
		packageName,
	)
	result := pattern.GetNamedStringFromText(protoContent, reg, "option_package_name")
	return result
}

func GetPackageNameFromProto(protoContent string) string {
	reg := "package[\\s]+(?P<package_name>[a-zA-Z0-9._\\-]+);"
	result := pattern.GetNamedStringFromText(protoContent, reg, "package_name")
	return result
}

func GetServiceNameFromProto(protoContent string) string {
	reg := "service[\\s]+(?P<service_name>[a-zA-Z0-9.]+)[\\s]*[\\\\r|\\\\n|]*[\\s]{"
	result := pattern.GetNamedStringFromText(protoContent, reg, "service_name")
	return result
}

func GenerateStub(
	content string,
	outputPath string,
	protocOpts []string,
	removeStubFiles bool,
) (map[string]string, error) {
	// add go_package in protobuf
	serviceName := GetServiceNameFromProto(content)
	protoFilePath := filepath.Join(outputPath, serviceName+".proto")
	io.SaveToFile(protoFilePath, []byte(content))

	// genereate protoc arguments
	opts := make([]string, 0)
	commonProtos := getCommonProtos()
	for _, opt := range commonProtos {
		opts = append(opts, fmt.Sprintf("-I=%s", opt))
	}
	opts = append(opts, protocOpts...)
	opts = append(opts, protoFilePath)

	// execute protoc
	logging.Debugf("exec cmd: %s\n",
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
	err = filepath.Walk(outputPath, func(path string, info os.FileInfo, err error) error {
		// read file path
		if info.IsDir() {
			return nil
		}
		relativePath := strings.TrimPrefix(path, outputPath)[1:]
		logging.Debugf("generated stub: %s\n", relativePath)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			logging.Errorf("failed to read %s\n", path)
			return err
		}
		stubs[relativePath] = string(data)
		return nil
	})

	if removeStubFiles {
		os.RemoveAll(outputPath)
		os.Remove(outputPath)
	}
	return stubs, err
}
