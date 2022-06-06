package pattern_test

import (
	"github.com/skema-dev/skema-tool/internal/pkg/pattern"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

func (s *testSuite) TestFindUrlPattern() {
	r := "https://github\\.com/(?P<organization_name>[a-zA-Z0-9-_]+)/(?P<repo_name>[a-zA-Z0-9-_]+)/(tree/main/){0,1}(?P<repo_path>[a-zA-Z0-9-_]+)"

	tests := []struct {
		content string
		pattern string
		name    string
		expect  string
	}{
		{
			content: "https://github.com/test-repo/ccc/tree/main/cc1/sds/sdd/pb",
			pattern: r,
			name:    "organization_name",
			expect:  "test-repo",
		},
		{
			content: "https://github.com/test-repo/ccc/tree/main/cc1/sds/sdd/pb",
			pattern: r,
			name:    "repo_name",
			expect:  "ccc",
		},
		{
			content: "https://github.com/test-repo/ccc/tree/main/cc1/sds/sdd/pb",
			pattern: r,
			name:    "repo_path",
			expect:  "cc1",
		},
		{
			content: "https://github.com/test-repo/ccc/cc1/sds/sdd/pb",
			pattern: r,
			name:    "repo_path",
			expect:  "cc1",
		},
		{
			content: "https://github.com/test-repo/ccc/cc1/sds/sdd/pb",
			pattern: r,
			name:    "organization_name",
			expect:  "test-repo",
		},
		{
			content: "https://github.com/test-repo/ccc/cc1/sds/sdd/pb",
			pattern: r,
			name:    "repo_name",
			expect:  "ccc",
		},
		{
			content: "https://github.com/12---/a_Z/---Z/sds/sdd/pb",
			pattern: r,
			name:    "repo_name",
			expect:  "a_Z",
		},
		{
			content: "https://github.com/12---/a_Z/---Z/sds/sdd/pb",
			pattern: r,
			name:    "repo_path",
			expect:  "---Z",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			result := pattern.GetNamedStringFromText(tt.content, tt.pattern, tt.name)
			assert.Equal(s.T(), tt.expect, result)
		})
	}

	found := pattern.GetNamedMapFromText("https://github.com/test-repo/ccc/tree/main/cc1/sds/sdd/pb", r, []string{"repo_name", "repo_path", "organization_name"})
	assert.Equal(s.T(), "test-repo", found["organization_name"])
	assert.Equal(s.T(), "ccc", found["repo_name"])
	assert.Equal(s.T(), "cc1", found["repo_path"])
}

func (s *testSuite) TestFindRepoPathPattern() {
	r := "https://github\\.com/(?P<organization_name>[a-zA-Z0-9-_]+)/(?P<repo_name>[a-zA-Z0-9-_]+)/(blob/main/){0,1}(?P<repo_path>[a-zA-Z0-9-_/.]+)"

	tests := []struct {
		content string
		pattern string
		name    string
		expect  string
	}{
		{
			content: "https://github.com/test-org/test-repo/blob/main/test004/abc/a123/test.proto",
			pattern: r,
			name:    "organization_name",
			expect:  "test-org",
		},
		{
			content: "https://github.com/test-org/test-repo/blob/main/test004/abc/a123/test.proto",
			pattern: r,
			name:    "repo_name",
			expect:  "test-repo",
		},
		{
			content: "https://github.com/test-org/test-repo/blob/main/test004/abc/a123/test.proto",
			pattern: r,
			name:    "repo_path",
			expect:  "test004/abc/a123/test.proto",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			result := pattern.GetNamedStringFromText(tt.content, tt.pattern, tt.name)
			assert.Equal(s.T(), tt.expect, result)
		})
	}
}
