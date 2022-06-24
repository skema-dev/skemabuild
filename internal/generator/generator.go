package generator

import (
	"bytes"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/skema-dev/skema-go/config"
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
		serviceTemplate *ServiceTemplate,
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
	serviceTemplate *ServiceTemplate,
) map[string]string {
	filepathPlaceHolders := make(map[string]string)
	filepathPlaceHolders["service_name"] = serviceTemplate.ProtocolServiceNameLower
	tpls := g.getTplContents(tpl)
	result := g.apply(tpls, serviceTemplate, filepathPlaceHolders)
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
		if err != nil {
			return err
		}

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
	serviceTemplate *ServiceTemplate,
	filepathPlaceholderNames map[string]string,
) map[string]string {
	result := make(map[string]string)
	console.Info("generating grpc service code...")

	templates := []string{"default.yaml", "default.yml"}
	for _, f := range templates {
		if content, ok := tpls[f]; ok {
			g.applyDefaultTemplateValue(content, serviceTemplate)
			break
		}
	}

	for tplFilepathName, tplContent := range tpls {
		filename := g.parseFilename(tplFilepathName, filepathPlaceholderNames)
		filename = strings.TrimSuffix(filename, ".tpl")

		if strings.HasSuffix(filename, "$model.go") {
			// special process for data models
			modelPath := filepath.Dir(filename)
			modelTpl := tplContent
			models := g.applyDataModel(modelPath, modelTpl, serviceTemplate.DataModels)
			for k, v := range models {
				result[k] = v
			}
			continue
		}

		content := g.parseTemplate(filename, tplContent, serviceTemplate)
		result[filename] = content
	}
	return result
}

func (g *grpcGoGenerator) applyDataModel(modelPath string, modelTpl string, dataModels []DataModelDescriptor) map[string]string {
	result := make(map[string]string)
	for _, model := range dataModels {
		filename := filepath.Join(modelPath, model.ModelNameLowerCase+".go")
		tpl := template.Must(template.New(filename).Option("missingkey=zero").Parse(modelTpl))
		var content bytes.Buffer
		err := tpl.Execute(&content, model)
		console.FatalIfError(err)
		result[filename] = content.String()
	}
	return result
}

func (g *grpcGoGenerator) parseTemplate(id string, tplContent string, serviceTemplate *ServiceTemplate) string {
	tpl := template.Must(template.New(id).Option("missingkey=zero").Parse(tplContent))
	var content bytes.Buffer
	err := tpl.Execute(&content, serviceTemplate)
	if err != nil {
		console.Info(err.Error())
		return tplContent
	}
	//console.FatalIfError(err)
	return content.String()
}

func (g *grpcGoGenerator) applyDefaultTemplateValue(templateYaml string, serviceTemplate *ServiceTemplate) {
	logging.Debugf("existing uservalues: %v", serviceTemplate.Value)
	conf := config.NewConfigWithString(templateYaml)

	if len(serviceTemplate.DataModels) == 0 {
		models := conf.GetStringArray("models")
		for _, v := range models {
			m := DataModelDescriptor{
				ModelNameCamelCase: strcase.ToCamel(v),
				ModelNameLowerCase: strings.ToLower(v),
			}
			console.Infof("add default data model: %s\n", m.ModelNameLowerCase)
			serviceTemplate.DataModels = append(serviceTemplate.DataModels, m)

			if serviceTemplate.DefaultDataModelNameLowerCase == "" {
				serviceTemplate.DefaultDataModelNameLowerCase = m.ModelNameLowerCase
				serviceTemplate.DefaultDataModelNameCamelCase = m.ModelNameCamelCase
			}
		}
	}

	values := conf.GetMapFromArray("values")
	for k, v := range values {
		k = strcase.ToCamel(k)
		if userValue, ok := serviceTemplate.Value[k]; !ok {
			serviceTemplate.Value[k] = v.(string)
			console.Infof("apply default template value: %s=>%s\n", k, v.(string))
		} else {
			console.Infof("template key: %s (%s), user defined: %s\n", k, v.(string), userValue)
		}
	}
}
