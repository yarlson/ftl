//go:build !race

package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/yarlson/ftl/pkg/config"
)

type DeploymentTestSuite struct {
	suite.Suite
	updater *Deployment
	network string
}

func TestDeploymentSuite(t *testing.T) {
	suite.Run(t, new(DeploymentTestSuite))
}

type LocalRunner struct{}

func (e *LocalRunner) RunCommand(ctx context.Context, command string, args ...string) (io.Reader, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	var combinedOutput bytes.Buffer
	cmd.Stdout = &combinedOutput
	cmd.Stderr = &combinedOutput

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return bytes.NewReader(combinedOutput.Bytes()), nil
}

func (e *LocalRunner) CopyFile(ctx context.Context, src, dst string) error {
	cmd := exec.CommandContext(ctx, "cp", src, dst)

	return cmd.Run()
}

func (suite *DeploymentTestSuite) SetupSuite() {
	suite.network = "ftl-test-network"
	_ = exec.Command("docker", "network", "create", suite.network).Run()
}

func (suite *DeploymentTestSuite) TearDownSuite() {
	_ = exec.Command("docker", "network", "rm", suite.network).Run()
}

func (suite *DeploymentTestSuite) SetupTest() {
	runner := &LocalRunner{}
	suite.updater = NewDeployment(runner, nil)
}

func (suite *DeploymentTestSuite) removeContainer(containerName string) {
	_ = exec.Command("docker", "stop", containerName).Run() // nolint: errcheck
	_ = exec.Command("docker", "rm", "-f", containerName).Run()
}

func (suite *DeploymentTestSuite) removeVolume(volumeName string) {
	_ = exec.Command("docker", "volume", "rm", volumeName).Run() // nolint: errcheck
}

func (suite *DeploymentTestSuite) inspectContainer(containerName string) map[string]interface{} {
	cmd := exec.Command("docker", "inspect", containerName)
	output, err := cmd.Output()
	assert.NoError(suite.T(), err)

	var containerInfo []map[string]interface{}
	err = json.Unmarshal(output, &containerInfo)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), containerInfo, 1)

	return containerInfo[0]
}

func (suite *DeploymentTestSuite) TestDeploy() {
	project := "test-project"

	cfg := &config.Config{
		Project: config.Project{
			Name:   project,
			Domain: "localhost",
			Email:  "test@example.com",
		},
		Services: []config.Service{
			{
				Name:  "web",
				Image: "nginx:1.19",
				Port:  80,
				Routes: []config.Route{
					{
						PathPrefix:  "/",
						StripPrefix: false,
					},
				},
			},
		},
		Dependencies: []config.Dependency{
			{
				Name:  "postgres",
				Image: "postgres:16",
				Volumes: []string{
					"postgres_data:/var/lib/postgresql/data",
				},
				EnvVars: map[string]string{
					"POSTGRES_PASSWORD": "S3cret",
					"POSTGRES_USER":     "test",
					"POSTGRES_DB":       "test",
				},
			},
			{
				Name:  "mysql",
				Image: "mysql:8",
				Volumes: []string{
					"mysql_data:/var/lib/mysql",
				},
				EnvVars: map[string]string{
					"MYSQL_ROOT_PASSWORD": "S3cret",
					"MYSQL_DATABASE":      "test",
					"MYSQL_USER":          "test",
					"MYSQL_PASSWORD":      "S3cret",
				},
			},
			{
				Name:  "mongodb",
				Image: "mongo:latest",
				Volumes: []string{
					"mongodb_data:/data/db",
				},
				EnvVars: map[string]string{
					"MONGO_INITDB_ROOT_USERNAME": "root",
					"MONGO_INITDB_ROOT_PASSWORD": "S3cret",
				},
			},
			{
				Name:  "redis",
				Image: "redis:latest",
				Volumes: []string{
					"redis_data:/data",
				},
			},
			{
				Name:  "rabbitmq",
				Image: "rabbitmq:management",
				Volumes: []string{
					"rabbitmq_data:/var/lib/rabbitmq",
				},
				EnvVars: map[string]string{
					"RABBITMQ_DEFAULT_USER": "user",
					"RABBITMQ_DEFAULT_PASS": "S3cret",
				},
			},
			{
				Name:  "elasticsearch",
				Image: "elasticsearch:7.14.0",
				Volumes: []string{
					"elasticsearch_data:/usr/share/elasticsearch/data",
				},
				EnvVars: map[string]string{
					"discovery.type": "single-node",
					"ES_JAVA_OPTS":   "-Xms512m -Xmx512m",
				},
			},
		},
		Volumes: []string{
			"postgres_data",
			"mysql_data",
			"mongodb_data",
			"redis_data",
			"rabbitmq_data",
			"elasticsearch_data",
		},
	}

	// Prepare the project folder
	projectPath, err := suite.updater.prepareProjectFolder(project)
	assert.NoError(suite.T(), err)

	// Generate certificates
	proxyCertPath := filepath.Join(projectPath, "localhost.crt")
	proxyKeyPath := filepath.Join(projectPath, "localhost.key")
	mkcertCmds := [][]string{
		{"mkcert", "-install"},
		{"mkcert", "-cert-file", proxyCertPath, "-key-file", proxyKeyPath, "localhost"},
	}

	for _, cmd := range mkcertCmds {
		output, err := suite.updater.runCommand(context.Background(), cmd[0], cmd[1:]...)
		assert.NoError(suite.T(), err, "Command output: %s", output)
	}

	// Cleanup containers and volumes before the test
	suite.cleanupDeployment()
	defer suite.cleanupDeployment()

	suite.Run("Successful deployment", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Deploy the initial configuration
		events := suite.updater.Deploy(ctx, project, cfg)
		for event := range events {
			suite.T().Logf("Event: %s", event)
			if event.Type == EventTypeError {
				assert.Fail(suite.T(), "Deployment error: %s", event.Message)
				return
			}
		}

		// Wait for services to stabilize
		time.Sleep(5 * time.Second)

		// Start making requests to test zero-downtime deployment
		var requestStats struct {
			totalRequests  int32
			failedRequests int32
		}

		requestCtx, requestCancel := context.WithCancel(context.Background())
		defer requestCancel()

		for i := 0; i < 10; i++ {
			go func() {
				for {
					select {
					case <-requestCtx.Done():
						return
					default:
						resp, err := http.Get("https://localhost/")
						atomic.AddInt32(&requestStats.totalRequests, 1)
						if err != nil || resp.StatusCode != http.StatusOK {
							atomic.AddInt32(&requestStats.failedRequests, 1)
						}
						if resp != nil {
							_ = resp.Body.Close()
						}
						time.Sleep(10 * time.Millisecond)
					}
				}
			}()
		}

		// Wait for some requests to be made
		time.Sleep(2 * time.Second)

		// Update the service image to trigger a redeployment
		cfg.Services[0].Image = "nginx:1.20"

		suite.T().Logf("Updating service image to nginx:1.20")

		// Redeploy with the updated configuration
		events = suite.updater.Deploy(ctx, project, cfg)
		for event := range events {
			suite.T().Logf("Event: %s", event)
			if event.Type == EventTypeError {
				assert.Fail(suite.T(), "Deployment error: %s", event.Message)
				return
			}
			// Optionally handle other events
		}

		// Wait for the redeployment to complete
		time.Sleep(2 * time.Second)
		requestCancel()

		fmt.Printf("Total requests: %d\n", requestStats.totalRequests)
		fmt.Printf("Failed requests: %d\n", requestStats.failedRequests)

		serviceName := "web"

		// Assert that there were no failed requests during the deployment
		assert.Equal(suite.T(), int32(0), requestStats.failedRequests, "Expected zero failed requests during zero-downtime deployment")

		containerInfo := suite.inspectContainer(serviceName)

		suite.Run("Updated Container State and Config", func() {
			state := containerInfo["State"].(map[string]interface{})
			config := containerInfo["Config"].(map[string]interface{})
			hostConfig := containerInfo["HostConfig"].(map[string]interface{})

			assert.Equal(suite.T(), "running", state["Status"])
			assert.Contains(suite.T(), config["Image"], "nginx:1.20")
			assert.Equal(suite.T(), project, hostConfig["NetworkMode"])
		})

		suite.Run("Updated Network Aliases", func() {
			networkSettings := containerInfo["NetworkSettings"].(map[string]interface{})
			networks := networkSettings["Networks"].(map[string]interface{})
			networkInfo := networks[project].(map[string]interface{})
			aliases := networkInfo["Aliases"].([]interface{})
			assert.Contains(suite.T(), aliases, serviceName)
		})
	})
}

// Helper function to clean up deployment artifacts
func (suite *DeploymentTestSuite) cleanupDeployment() {
	containers := []string{"proxy", "web", "postgres", "mysql", "mongodb", "redis", "rabbitmq", "elasticsearch", "certrenewer"}
	volumes := []string{
		"test-project-postgres_data",
		"test-project-mysql_data",
		"test-project-mongodb_data",
		"test-project-redis_data",
		"test-project-rabbitmq_data",
		"test-project-elasticsearch_data",
	}

	for _, container := range containers {
		suite.removeContainer(container)
	}

	for _, volume := range volumes {
		suite.removeVolume(volume)
	}
}
