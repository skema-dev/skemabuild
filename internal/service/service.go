package service

import (
	"github.com/google/uuid"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/io"
	"github.com/skema-dev/skema-tool/internal/pkg/pattern"
	"google.golang.org/protobuf/types/descriptorpb"
	"os"
	"path/filepath"
	"strings"
)

func GetRemoteProtobufLocation(url string) (string, string) {
	r := "https://github\\.com/(?P<organization_name>[a-zA-Z0-9-_]+)/(?P<repo_name>[a-zA-Z0-9-_]+)/(blob/main/){0,1}(?P<repo_path>[a-zA-Z0-9-_/.]+)"
	resourceMap := pattern.GetNamedMapFromText(url, r, []string{"repo_name", "repo_path"})
	repoName := resourceMap["repo_name"]
	repoPath := resourceMap["repo_path"]
	return repoName, repoPath
}

func GetProtobufDescriptionFromString(content string, importPaths ...string) []*descriptorpb.FileDescriptorProto {

	tempFilePath := filepath.Join(io.GetHomePath(), "temp", strings.ReplaceAll(uuid.New().String(), "-", ""))
	io.SaveToFile(tempFilePath, []byte(content))
	defer os.Remove(tempFilePath)

	parser := protoparse.Parser{
		ImportPaths:           importPaths,
		IncludeSourceCodeInfo: true,
	}

	descriptors, err := parser.ParseFilesButDoNotLink(tempFilePath)
	if err != nil {
		console.Fatalf(err.Error())
	}

	console.Infof("%d\n", len(descriptors))

	return descriptors
}
