package service

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/io"
	"github.com/skema-dev/skema-tool/internal/pkg/pattern"
	"google.golang.org/protobuf/types/descriptorpb"
)

func GetGithubContentLocation(url string) (string, string, string) {
	r := "https://github\\.com/(?P<organization_name>[a-zA-Z0-9-_]+)/(?P<repo_name>[a-zA-Z0-9-_]+)/(blob/main/){0,1}(tree/main/){0,1}(?P<repo_path>[a-zA-Z0-9-_/.]+)"
	resourceMap := pattern.GetNamedMapFromText(url, r, []string{"organization_name", "repo_name", "repo_path"})
	repoName := resourceMap["repo_name"]
	repoPath := resourceMap["repo_path"]
	organization := resourceMap["organization_name"]
	return repoName, repoPath, organization
}

func GetRpcParameters(
	protoContent string,
	goModule string,
	goVersion string,
	userServiceName string,
) *RpcParameters {
	parameters := &RpcParameters{}
	parameters.GoModule = goModule
	parameters.GoVersion = goVersion
	parameters.ServiceName = userServiceName

	descriptor := getProtobufFileDescriptor(protoContent)
	parameters.GoPackageAddress = getOptionPackageAddress(descriptor)
	if parameters.GoPackageAddress == "" {
		console.Fatalf("incorrect go package address. please check the proto file.")
	}

	for _, s := range descriptor.Service {
		serviceDescriptor := ServiceDescriptor{}
		serviceDescriptor.Name = s.GetName()

		if parameters.ServiceName == "" {
			parameters.ServiceName = serviceDescriptor.Name + "Service"
		}
		parameters.ServiceNameCamelCase = strcase.ToCamel(parameters.ServiceName)
		parameters.ServiceNameLower = strings.ToLower(parameters.ServiceName)

		if parameters.GoModule == "" {
			parameters.GoModule = parameters.ServiceName
		}

		for _, m := range s.GetMethod() {
			methodDescriptor := ServiceMethodDescriptor{}
			methodDescriptor.Name = m.GetName()
			methodDescriptor.NameCamelCase = strcase.ToLowerCamel(methodDescriptor.Name)
			methodDescriptor.RequestType = m.GetInputType()
			methodDescriptor.ResponseType = m.GetOutputType()
			serviceDescriptor.Methods = append(serviceDescriptor.Methods, methodDescriptor)
		}

		parameters.RpcServices = append(parameters.RpcServices, serviceDescriptor)
	}

	console.Info(
		"go module: %s\ngo version: %s\nservice: %s\ngo package: %s\n",
		parameters.GoModule,
		parameters.GoVersion,
		parameters.ServiceName,
		parameters.GoPackageAddress,
	)

	return parameters
}

func getProtobufFileDescriptor(content string) *descriptorpb.FileDescriptorProto {
	tempFilePath := filepath.Join(
		io.GetHomePath(),
		"temp",
		strings.ReplaceAll(uuid.New().String(), "-", ""),
	)
	io.SaveToFile(tempFilePath, []byte(content))
	defer os.Remove(tempFilePath)

	parser := protoparse.Parser{
		ImportPaths:           []string{},
		IncludeSourceCodeInfo: true,
	}
	descriptors, err := parser.ParseFilesButDoNotLink(tempFilePath)
	if err != nil {
		console.Fatalf(err.Error())
	}
	if len(descriptors) > 1 {
		console.Fatalf("multiple service descriptors found. ")
	}

	return descriptors[0]
}

func getOptionPackageAddress(fd *descriptorpb.FileDescriptorProto) string {
	if fd == nil || fd.Options == nil {
		return ""
	}

	uninterpreted := fd.Options.UninterpretedOption
	for _, option := range uninterpreted {
		name := option.Name
		for _, p := range name {
			switch *p.NamePart {
			case "go_package":
				fallthrough
			case "java_package":
				return string(option.StringValue)
			}
		}
	}
	return ""
}
