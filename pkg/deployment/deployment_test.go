//go:build !race

package deployment

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/runner/remote"
	"github.com/yarlson/ftl/pkg/ssh"
	"github.com/yarlson/ftl/tests/dockercontainer"
)

type DeploymentTestSuite struct {
	suite.Suite
	runner     *remote.Runner
	deployment *Deployment
	network    string
	tc         *dockercontainer.Container
}

func TestDeploymentSuite(t *testing.T) {
	suite.Run(t, new(DeploymentTestSuite))
}

func (suite *DeploymentTestSuite) SetupTest() {
	suite.T().Log("Setting up test environment...")
	tc, err := dockercontainer.NewContainer(suite.T())
	suite.Require().NoError(err)

	suite.network = "ftl-test-network"
	suite.tc = tc

	suite.T().Log("Creating SSH client...")
	sshClient, err := ssh.NewSSHClientWithPassword("127.0.0.1", tc.SshPort.Port(), "root", "testpassword")
	suite.Require().NoError(err)

	suite.T().Log("Creating runner...")
	runner := remote.NewRunner(sshClient)
	suite.runner = runner
	suite.deployment = NewDeployment(runner, nil)
}

func (suite *DeploymentTestSuite) TearDownTest() {
	_ = suite.tc.Container.Terminate(context.Background())
}

func (suite *DeploymentTestSuite) removeContainer(containerName string) {
	_, _ = suite.runner.RunCommand(context.Background(), "docker", "stop", containerName)
	_, _ = suite.runner.RunCommand(context.Background(), "docker", "rm", "-f", containerName)
}

func (suite *DeploymentTestSuite) removeVolume(volumeName string) {
	_, _ = suite.runner.RunCommand(context.Background(), "docker", "volume", "rm", volumeName)
}

func (suite *DeploymentTestSuite) inspectContainer(containerName string) map[string]interface{} {
	output, err := suite.runner.RunCommand(context.Background(), "docker", "inspect", containerName)
	suite.Require().NoError(err)
	outputBytes, err := io.ReadAll(output)
	suite.Require().NoError(err)

	var containerInfo []map[string]interface{}
	err = json.Unmarshal(outputBytes, &containerInfo)
	suite.Require().NoError(err)
	suite.Require().Len(containerInfo, 1)

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
		},
		Volumes: []string{
			"postgres_data",
		},
	}

	projectPath, err := suite.deployment.prepareProjectFolder(project)
	suite.Require().NoError(err)

	proxyCertPath := filepath.Join(projectPath, "localhost.crt")
	proxyKeyPath := filepath.Join(projectPath, "localhost.key")
	mkcertCmds := [][]string{
		{"mkcert", "-install"},
		{"mkcert", "-cert-file", proxyCertPath, "-key-file", proxyKeyPath, "localhost"},
	}

	for _, cmd := range mkcertCmds {
		output, err := suite.deployment.runCommand(context.Background(), cmd[0], cmd[1:]...)
		suite.Require().NoError(err, "Command output: %s", output)
	}

	suite.cleanupDeployment()
	defer suite.cleanupDeployment()

	suite.Run("Successful deployment", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		events := suite.deployment.Deploy(ctx, project, cfg)
		for event := range events {
			suite.T().Logf("Event: %s", event)
			if event.Type == EventTypeError {
				suite.Require().Fail(event.Message, "Deployment error %s", event.Message)
				return
			}
		}

		time.Sleep(5 * time.Second)

		var requestStats struct {
			totalRequests  int32
			failedRequests int32
		}

		requestCtx, requestCancel := context.WithCancel(context.Background())
		defer requestCancel()

		for i := 0; i < 10; i++ {
			go func() {
				client := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					},
				}
				for {
					select {
					case <-requestCtx.Done():
						return
					default:
						port := suite.tc.SslPort.Port()
						resp, err := client.Get(fmt.Sprintf("https://localhost:%s/", port))
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

		time.Sleep(2 * time.Second)

		cfg.Services[0].Image = "nginx:1.20"

		suite.T().Logf("Updating service image to nginx:1.20")

		events = suite.deployment.Deploy(ctx, project, cfg)
		for event := range events {
			suite.T().Logf("Event: %s", event)
			if event.Type == EventTypeError {
				suite.Require().Fail(event.Message, "Deployment error %s", event.Message)
				return
			}
		}

		time.Sleep(2 * time.Second)
		requestCancel()

		fmt.Printf("Total requests: %d\n", requestStats.totalRequests)
		fmt.Printf("Failed requests: %d\n", requestStats.failedRequests)

		serviceName := "web"

		suite.Require().Equal(int32(0), requestStats.failedRequests, "Expected zero failed requests during zero-downtime deployment")

		containerInfo := suite.inspectContainer(serviceName)

		suite.Run("Updated Container State and Config", func() {
			state := containerInfo["State"].(map[string]interface{})
			cfg := containerInfo["Config"].(map[string]interface{})
			hostConfig := containerInfo["HostConfig"].(map[string]interface{})

			suite.Require().Equal("running", state["Status"])
			suite.Require().Contains(cfg["Image"], "nginx:1.20")
			suite.Require().Equal(project, hostConfig["NetworkMode"])
		})

		suite.Run("Updated Network Aliases", func() {
			networkSettings := containerInfo["NetworkSettings"].(map[string]interface{})
			networks := networkSettings["Networks"].(map[string]interface{})
			networkInfo := networks[project].(map[string]interface{})
			aliases := networkInfo["Aliases"].([]interface{})
			suite.Require().Contains(aliases, serviceName)
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
