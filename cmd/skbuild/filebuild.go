package main

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
	"github.com/skema-dev/skemabuild/cmd/skbuild/api"
	"github.com/skema-dev/skemabuild/cmd/skbuild/service"
	internalApi "github.com/skema-dev/skemabuild/internal/api"
	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/skema-dev/skemabuild/internal/pkg/io"
	"strings"
)

func buildServiceFromConfigFile(configFilepath string, username string, password string) {
	logging.Init("info", "console")
	conf := config.NewConfigWithFile(configFilepath)
	if conf == nil {
		console.Fatalf("Failed to load config from %s\n", configFilepath)
	}
	protoFilepath := conf.GetString("api.proto")
	protoContent := io.GetContentFromUri(protoFilepath)

	goPackage := internalApi.GetOptionGoPackageNameFromProto(protoContent)
	if goPackage == "" {
		console.Fatalf(`
						No go_package option defined. Please add go_package in your proto file
						example:
						option go_package="github.com/likezhang-public/newst001/test001/com.pack1/grpc-go";
						`)
	}
	stubs := conf.GetStringArray("api.stubs")
	stubList := strings.Join(stubs, ",")
	version := conf.GetString("api.version", "")
	uploadPath := conf.GetString("api.path", "")
	if uploadPath != "" && version != "" {
		api.PublishFromProto(protoContent, stubList, uploadPath, version, username, password)
	}

	serviceParams := ParseServiceConfig(conf)
	console.Info("stub published. Creating service code...")
	service.CreateServiceCode(serviceParams)
}

func ParseServiceConfig(conf *config.Config) *service.ServiceGeneratorParameters {
	serviceParams := service.ServiceGeneratorParameters{
		Models: make([]string, 0),
		Values: make(map[string]string),
	}
	if v := conf.GetString("service.name", ""); v != "" {
		serviceParams.ServiceName = v
	}
	serviceParams.Models = conf.GetStringArray("service.template.models")
	userValues := conf.GetMapFromArray("service.template.values")
	for k, v := range userValues {
		serviceParams.Values[strcase.ToCamel(k)] = fmt.Sprintf("%v", v)
	}

	serviceParams.ProtoUri = conf.GetString("api.proto")
	serviceParams.Tpl = conf.GetString("service.tpl", "skema-complete")
	serviceParams.HttpEnabled = conf.GetBool("service.http_enabled", true)
	serviceParams.GoVersion = conf.GetString("go_version", "1.16")
	serviceParams.GoModule = serviceParams.ServiceName

	return &serviceParams
}
