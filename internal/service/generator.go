package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/skema-dev/skema-go/logging"
	"github.com/skema-dev/skemabuild/internal/auth"
	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/skema-dev/skemabuild/internal/pkg/pattern"
	"github.com/skema-dev/skemabuild/internal/pkg/repository"
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
	filepathPlaceHolders := make(map[string]string)
	filepathPlaceHolders["service_name"] = rpcParameters.ServiceNameLower
	tpls := g.getTplContents(tpl)
	result := g.apply(tpls, rpcParameters, filepathPlaceHolders, userParameters)
	return result
}

func (g *grpcGoGenerator) getTplContents(tpl string) map[string]string {
	tplPath := tpl
	console.Info("get code template from " + tplPath)

	// read local templates
	if strings.HasPrefix(tplPath, "file://") {
		tplPath = strings.TrimPrefix(tplPath, "file://")
		if strings.HasPrefix(tplPath, "~/") {
			tplPath = strings.TrimPrefix(tplPath, "~/")
			dirname, err := os.UserHomeDir()
			console.FatalIfError(err)
			tplPath = filepath.Join(dirname, tplPath)
		}
		tpls := g.getLocalTpls(tplPath)
		return tpls
	}

	fromSkemaTemplateRepo := false
	if !pattern.IsHttpUrl(tplPath) {
		// shortcut name for skema template repository
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

func (g *grpcGoGenerator) getLocalTpls(startPath string) map[string]string {
	tpls := make(map[string]string)
	err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		// read file path
		console.Info(path)
		console.FatalIfError(err)
		if info == nil || info.IsDir() {
			return nil
		}
		relativePath := strings.TrimPrefix(path, startPath)[1:]
		logging.Debugf("tpl file: %s\n", relativePath)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			logging.Errorf("failed to read %s\n", path)
			return err
		}
		tpls[relativePath] = string(data)
		return nil
	})
	console.FatalIfError(err)

	return tpls
}

func (g *grpcGoGenerator) parseFilename(
	tplFilepathName string,
	filepathPlaceholderNames map[string]string,
) string {
	placeHolderStart := strings.Index(tplFilepathName, "$")
	placeHolderEnd := strings.Index(tplFilepathName, "#")
	if placeHolderStart < 0 || placeHolderEnd <= placeHolderStart {
		return tplFilepathName
	}

	placeholder := tplFilepathName[placeHolderStart+1 : placeHolderEnd]
	k := placeholder
	v, ok := filepathPlaceholderNames[k]
	if !ok {
		console.Fatalf("missing filename placeholder %s", k)
	}
	filename := fmt.Sprintf(
		"%s%s%s",
		tplFilepathName[:placeHolderStart],
		v,
		tplFilepathName[placeHolderEnd+1:len(tplFilepathName)],
	)
	return filename
}

func (g *grpcGoGenerator) apply(
	tpls map[string]string,
	parameters *RpcParameters,
	filepathPlaceholderNames map[string]string,
	userParameters map[string]string,
) map[string]string {
	result := make(map[string]string)
	console.Info("generating grpc service code...")
	for tplFilepathName, tplContent := range tpls {
		filename := g.parseFilename(tplFilepathName, filepathPlaceholderNames)
		filename = strings.TrimSuffix(filename, ".tpl")

		tpl := template.Must(template.New(filename).Option("missingkey=zero").Parse(tplContent))
		var content bytes.Buffer
		err := tpl.Execute(&content, parameters)
		console.FatalIfError(err)

		var contentFinal bytes.Buffer
		tplFinal := template.Must(template.New(filename + "_final").Option("missingkey=zero").Parse(content.String()))
		err = tplFinal.Execute(&contentFinal, userParameters)
		console.FatalIfError(err)

		result[filename] = contentFinal.String()
	}
	return result
}
