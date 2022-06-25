package generator

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/skema-dev/skemabuild/internal/auth"
	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/skema-dev/skemabuild/internal/pkg/http"
	"github.com/skema-dev/skemabuild/internal/pkg/io"
	"github.com/skema-dev/skemabuild/internal/pkg/pattern"
	"github.com/skema-dev/skemabuild/internal/pkg/repository"
	"google.golang.org/protobuf/types/descriptorpb"
	"os"
	"path/filepath"
	"strings"
)

type ServiceMethodDescriptor struct {
	Name          string
	NameCamelCase string
	RequestType   string
	ResponseType  string
}

type ServiceDescriptor struct {
	Name    string
	Methods []ServiceMethodDescriptor
}

type ServiceTemplate struct {
	GoModule                 string
	GoVersion                string
	GoPackageAddress         string
	HttpEnabled              bool
	ProtocolServiceName      string
	ProtocolServiceNameLower string
	ServiceName              string
	ServiceNameCamelCase     string
	ServiceNameLower         string

	DefaultDataModelNameCamelCase string
	DefaultDataModelNameLowerCase string

	RpcServices []ServiceDescriptor

	DataModels []DataModelDescriptor

	Value map[string]string
}

type DataModelDescriptor struct {
	ModelNameCamelCase string
	ModelNameLowerCase string
}

func CreateServiceTemplate() *ServiceTemplate {
	serviceTemplate := &ServiceTemplate{}
	return serviceTemplate
}

func (t *ServiceTemplate) WithRpcProtocol(protoUri string, goModule string, goVersion string, serviceName string, httpEnabled bool) *ServiceTemplate {
	if pattern.IsGithubUrl(protoUri) {
		// use github client to get proto file
		authProvider := auth.NewGithubAuthProvider()
		repo := repository.NewGithubRepo(authProvider.GetLocalToken())
		if repo == nil {
			console.Fatalf("failed to initiate github repo")
		}
		repoName, repoPath, _ := GetGithubContentLocation(protoUri)
		console.Info("get remote proto on github: %s", protoUri)
		console.Info("Repo: %s\nPath: %s", repoName, repoPath)

		content, err := repo.GetContents(repoName, repoPath)
		if err != nil {
			console.Fatalf(err.Error())
		}
		t = t.WithRpcParameters(
			content[repoPath],
			goModule,
			goVersion,
			serviceName,
		)
	} else if pattern.IsHttpUrl(protoUri) {
		// get proto by regular http
		console.Info("get remote proto: %s", protoUri)
		content := http.GetTextContent(protoUri)
		t = t.WithRpcParameters(content, goModule, goVersion, serviceName)
	} else {
		// read from local path
		data, err := os.ReadFile(protoUri)
		console.FatalIfError(err, fmt.Sprintf("Failed reading proto from \"%s\" ", protoUri))
		content := string(data)
		t = t.WithRpcParameters(content, goModule, goVersion, serviceName)
	}
	t.HttpEnabled = httpEnabled

	return t
}

func (t *ServiceTemplate) WithRpcParameters(
	protoContent string,
	goModule string,
	goVersion string,
	userServiceName string,
) *ServiceTemplate {
	t.GoModule = goModule
	t.GoVersion = goVersion
	t.ProtocolServiceName = userServiceName
	t.ProtocolServiceNameLower = strings.ToLower(userServiceName)

	descriptor := t.getProtobufFileDescriptor(protoContent)
	t.GoPackageAddress = t.getOptionPackageAddress(descriptor)
	if t.GoPackageAddress == "" {
		console.Fatalf("incorrect go package address. please check the proto file.")
	}

	for _, s := range descriptor.Service {
		serviceDescriptor := ServiceDescriptor{}
		serviceDescriptor.Name = s.GetName()

		if t.ServiceName == "" {
			t.ServiceName = serviceDescriptor.Name + "Service"
			t.ServiceNameCamelCase = strcase.ToCamel(t.ServiceName)
			t.ServiceNameLower = strings.ToLower(t.ServiceName)
		}

		if t.ProtocolServiceName == "" {
			t.ProtocolServiceName = serviceDescriptor.Name
			t.ProtocolServiceNameLower = strings.ToLower(serviceDescriptor.Name)
		}

		if t.GoModule == "" {
			t.GoModule = t.ServiceName
		}

		for _, m := range s.GetMethod() {
			methodDescriptor := ServiceMethodDescriptor{}
			methodDescriptor.Name = m.GetName()
			methodDescriptor.NameCamelCase = strcase.ToLowerCamel(methodDescriptor.Name)
			methodDescriptor.RequestType = m.GetInputType()
			methodDescriptor.ResponseType = m.GetOutputType()
			serviceDescriptor.Methods = append(serviceDescriptor.Methods, methodDescriptor)
		}

		t.RpcServices = append(t.RpcServices, serviceDescriptor)
	}

	console.Info(
		"go module: %s\ngo version: %s\nservice: %s\ngo package: %s\n",
		t.GoModule,
		t.GoVersion,
		t.ServiceName,
		t.GoPackageAddress,
	)

	return t
}

func (t *ServiceTemplate) getProtobufFileDescriptor(content string) *descriptorpb.FileDescriptorProto {
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

func (t *ServiceTemplate) getOptionPackageAddress(fd *descriptorpb.FileDescriptorProto) string {
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

func (t *ServiceTemplate) WithDataModelNames(modelNames []string) *ServiceTemplate {
	descriptors := make([]DataModelDescriptor, 0)
	for _, name := range modelNames {
		desc := DataModelDescriptor{
			ModelNameCamelCase: strcase.ToCamel(name),
			ModelNameLowerCase: strings.ToLower(name),
		}

		descriptors = append(descriptors, desc)
	}

	if len(descriptors) > 0 {
		t.DefaultDataModelNameCamelCase = descriptors[0].ModelNameCamelCase
		t.DefaultDataModelNameLowerCase = descriptors[0].ModelNameLowerCase
	}

	t.DataModels = descriptors
	return t
}

func (t *ServiceTemplate) WithUserValues(values map[string]string) *ServiceTemplate {
	t.Value = values
	return t
}
