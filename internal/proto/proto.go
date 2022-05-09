package proto

import (
	"bytes"
	_ "embed"
	"github.com/iancoleman/strcase"
	"strings"
	"text/template"
)

//go:embed tpl/default_pb.tpl
var protobufTemplate string

type Proto struct {
}

func (p *Proto) InitProtoFile(moduleName string, packageName string, serviceName string, options []string) (string, error) {
	tpl := template.Must(template.New("default").Parse(protobufTemplate))

	opts := ""
	for _, v := range options {
		if v == "" {
			continue
		}
		opts = opts + "option " + v + ";\n"
	}

	userValues := map[string]string{
		"Module":  strings.ToLower(moduleName),
		"Package": strings.ToLower(packageName),
		"Service": strcase.ToCamel(serviceName),
		"Options": opts,
	}

	var content bytes.Buffer
	err := tpl.Execute(&content, userValues)
	if err != nil {
		return "", err
	}

	return content.String(), nil
}
