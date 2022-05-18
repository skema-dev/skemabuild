package api

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

//go:embed tpl/default_pb.tpl
var protobufTemplate string

type ApiCreator interface {
	InitProtoFile(packageName string, serviceName string, options []string) (string, error)
}

func NewApiCreator() ApiCreator {
	return &creator{}
}

type creator struct {
}

func (p *creator) InitProtoFile(
	packageName string,
	serviceName string,
	options []string,
) (string, error) {
	tpl := template.Must(template.New("default").Option("missingkey=zero").Parse(protobufTemplate))

	opts := ""
	for _, v := range options {
		if v == "" {
			continue
		}
		opts = opts + "option " + v + ";\n"
	}

	userValues := map[string]string{
		"Package": strings.ToLower(packageName),
		"Service": strcase.ToCamel(serviceName),
		"Options": opts,
	}

	if opts == "" {
		userValues["Options"] = ProtocobufOptionTplStr
	}

	var content bytes.Buffer
	err := tpl.Execute(&content, userValues)
	if err != nil {
		return "", err
	}

	return content.String(), nil
}
