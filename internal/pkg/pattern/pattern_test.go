package pattern_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"skema-tool/internal/pkg/pattern"
	"testing"
)

type testSuite struct {
	suite.Suite
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestFindNamedPattern() {
	content := "abc 123 \n package    option=\"abc.123\";\n   service     Test1Service \n{}"

	tests := []struct {
		pattern string
		name    string
		expect  string
	}{
		{
			pattern: "package[\\s]+option=\"(?P<option_name>[a-zA-Z0-9.]+)\";",
			name:    "option_name",
			expect:  "abc.123",
		},
		{
			pattern: "service[\\s]+(?P<service_name>[a-zA-Z0-9.]+)[\\s]*[\\r\\n]+",
			name:    "service_name",
			expect:  "Test1Service",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			result := pattern.GetNamedStringFromText(content, tt.pattern, tt.name)
			assert.Equal(s.T(), tt.expect, result)
		})
	}
}
