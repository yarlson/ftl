package proxy

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/yarlson/ftl/pkg/config"
)

type ProxyTestSuite struct {
	suite.Suite
}

func TestProxySuite(t *testing.T) {
	suite.Run(t, new(ProxyTestSuite))
}

func stripWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
}

func (suite *ProxyTestSuite) TestGenerateNginxConfig_Success() {
	cfg := &config.Config{
		Project: config.Project{
			Name:   "test-project",
			Domain: "test.example.com",
			Email:  "test@example.com",
		},
		Services: []config.Service{
			{
				Name:  "web",
				Image: "nginx:latest",
				Port:  80,
				Routes: []config.Route{
					{
						PathPrefix:  "/",
						StripPrefix: false,
					},
				},
			},
		},
	}

	_, err := GenerateNginxConfig(cfg)
	assert.NoError(suite.T(), err)
}

func (suite *ProxyTestSuite) TestGenerateNginxConfig_MultipleServices() {
	cfg := &config.Config{
		Project: config.Project{
			Name:   "test-project",
			Domain: "test.example.com",
			Email:  "test@example.com",
		},
		Services: []config.Service{
			{
				Name:  "web",
				Image: "nginx:latest",
				Port:  80,
				Routes: []config.Route{
					{
						PathPrefix:  "/",
						StripPrefix: false,
					},
				},
			},
			{
				Name:  "api",
				Image: "api:latest",
				Port:  8080,
				Routes: []config.Route{
					{
						PathPrefix:  "/api",
						StripPrefix: true,
					},
				},
			},
		},
	}

	_, err := GenerateNginxConfig(cfg)

	assert.NoError(suite.T(), err)
}

func (suite *ProxyTestSuite) TestGenerateNginxConfig_EmptyConfig() {
	cfg := &config.Config{}

	_, err := GenerateNginxConfig(cfg)

	assert.NoError(suite.T(), err)
}
