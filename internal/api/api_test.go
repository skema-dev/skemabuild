package api_test

import (
	"github.com/stretchr/testify/assert"
	"skema-tool/internal/api"
	"strings"
	"testing"
)

func TestCreateAPIandStub(t *testing.T) {
	apiCreator := api.NewApiCreator()
	stubCreator := api.NewGoStubCreator("github.com/test/abc123")

	packageName := "abc.abc"
	serviceName := "test1"

	protoContent, err := apiCreator.InitProtoFile(packageName, serviceName, make([]string, 0))
	assert.Nil(t, err)
	stubs, err := stubCreator.Generate(protoContent)
	assert.Nil(t, err)

	assert.True(t, len(stubs) > 1)
	foundGateway := false
	foundValidator := false
	foundGrpcPb := false
	for k, _ := range stubs {
		if strings.HasSuffix(k, ".pb.gw.go") {
			foundGateway = true
		} else if strings.HasSuffix(k, ".pb.validate.go") {
			foundValidator = true
		} else if strings.HasSuffix(k, "_grpc.pb.go") {
			foundGrpcPb = true
		}
	}

	assert.True(t, foundGateway)
	assert.True(t, foundValidator)
	assert.True(t, foundGrpcPb)
}
