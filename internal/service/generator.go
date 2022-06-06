package service

import (
	"bytes"
	"fmt"
	"github.com/skema-dev/skema-tool/internal/auth"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/repository"
	"strings"
	"text/template"
)

type Generator interface {
	CreateCodeContent(
		tpl string,
		rpcParameters *RpcParameters,
		userParameters map[string]string,
	) map[string]string
}

type grpcGoGenerator struct {
}

func NewGrpcGoGenerator() Generator {
	g := &grpcGoGenerator{}
	return g
}

func (g *grpcGoGenerator) CreateCodeContent(
	tpl string,
	rpcParameters *RpcParameters,
	userParameters map[string]string,
) map[string]string {
	userParameters["filename_placeholder_service_name"] = strings.ToLower(rpcParameters.ServiceName)
	tpls := g.getTplContents(tpl)
	result := g.apply(tpls, rpcParameters, userParameters)
	return result
}

func (g *grpcGoGenerator) getTplContents(tpl string) map[string]string {
	tplPath := tpl
	fromSkemaTemplateRepo := false
	if !strings.HasPrefix(tpl, "https://") && !strings.HasPrefix(tpl, "http://") {
		defaultHostPath := "https://github.com/skema-dev/template/grpc-go"
		tplPath = fmt.Sprintf("%s/%s", defaultHostPath, tpl)
		fromSkemaTemplateRepo = true
	}
	authProvider := auth.NewGithubAuthProvider()
	repo := repository.NewGithubRepo(authProvider.GetLocalToken())
	repoName, repoPath, organization := GetGithubContentLocation(tplPath)
	tpls, err := repo.GetContents(repoName, repoPath, organization)
	if err != nil {
		console.Fatalf("tpl url: %s\n%s", tplPath, err.Error())
	}

	if fromSkemaTemplateRepo {
		newTpls := make(map[string]string)
		prefix := fmt.Sprintf("grpc-go/%s/", tpl)
		for k, v := range tpls {
			newKey := strings.TrimPrefix(k, prefix)
			newTpls[newKey] = v
		}
		return newTpls
	}

	return tpls
}

func (g *grpcGoGenerator) parseFilename(tplFilepathName string, userParameters map[string]string) string {
	placeHolderStart := strings.Index(tplFilepathName, "$")
	placeHolderEnd := strings.Index(tplFilepathName, "#")
	if placeHolderStart < 0 || placeHolderEnd <= placeHolderStart {
		return tplFilepathName
	}

	placeholder := tplFilepathName[placeHolderStart+1 : placeHolderEnd]
	k := fmt.Sprintf("filename_placeholder_%s", placeholder)
	v, ok := userParameters[k]
	if !ok {
		console.Fatalf("missing filename placeholder %s", k)
	}
	filename := fmt.Sprintf("%s%s%s", tplFilepathName[:placeHolderStart], v, tplFilepathName[placeHolderEnd+1:len(tplFilepathName)])
	return filename
}

func (g *grpcGoGenerator) apply(tpls map[string]string, parameters *RpcParameters, userParameters map[string]string) map[string]string {
	result := make(map[string]string)
	for tplFilepathName, tplContent := range tpls {
		filename := g.parseFilename(tplFilepathName, userParameters)
		filename = strings.TrimSuffix(filename, ".tpl")

		tpl := template.Must(template.New(filename).Option("missingkey=zero").Parse(tplContent))
		var content bytes.Buffer
		err := tpl.Execute(&content, parameters)
		if err != nil {
			console.Fatalf(err.Error())
		}
		result[filename] = content.String()
	}
	return result
}
