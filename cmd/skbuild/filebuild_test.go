package main

import (
	"github.com/skema-dev/skema-go/config"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	suite.Suite
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestParseConfig() {
	testContent := `
api:
  proto: test1.proto
  stubs: 
    - grpc-go
    - openapi
  version: v0.0.1
  path: ./

service:
  name: mytestservice
  tpl: skema-complete
  template:
    models:
      - user
      - message      
    values:
      - mysql_server_name: mysql-svc
      - mysql_password: abcd1234
      - grpc_port: "9991"
      - http_port: "9992"
`
	conf := config.NewConfigWithString(testContent)
	protoFilepath := conf.GetString("api.proto")
	stubs := conf.GetStringArray("api.stubs")
	stubList := strings.Join(stubs, ",")
	version := conf.GetString("api.version")
	uploadPath := conf.GetString("api.path")
	serviceParams := ParseServiceConfig(conf)

	assert.Equal(s.T(), protoFilepath, "test1.proto")
	assert.Equal(s.T(), stubList, "grpc-go,openapi")
	assert.Equal(s.T(), version, "v0.0.1")
	assert.Equal(s.T(), uploadPath, "./")

	assert.Equal(s.T(), serviceParams.ServiceName, "mytestservice")
	assert.Equal(s.T(), len(serviceParams.Models), 2)
	assert.Equal(s.T(), serviceParams.Tpl, "skema-complete")
	assert.Equal(s.T(), len(serviceParams.Values), 4)
}
