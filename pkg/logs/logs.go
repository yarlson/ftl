package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/deployment"
)

// Logger provides methods to fetch logs from remote services.
type Logger struct {
	runner deployment.Runner
}

// NewLogger creates a new Logger instance.
func NewLogger(runner deployment.Runner) *Logger {
	return &Logger{runner: runner}
}

// FetchLogs fetches and prints logs from the specified service.
func (l *Logger) FetchLogs(ctx context.Context, project, service string) error {
	containerName := service

	// Check if the container exists
	exists, err := l.containerExists(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to check if container exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("service %s is not running on the server", service)
	}

	// Run the docker logs command
	reader, err := l.runner.RunCommand(ctx, "docker", "logs", containerName)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		console.Info(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
}

// StreamLogs streams logs from the specified service in real-time.
func (l *Logger) StreamLogs(ctx context.Context, project, service string) error {
	containerName := service

	// Check if the container exists
	exists, err := l.containerExists(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to check if container exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("service %s is not running on the server", service)
	}

	// Run the docker logs -f command
	reader, err := l.runner.RunCommand(ctx, "docker", "logs", "-f", containerName)
	if err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	// Stream the logs in real-time
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		console.Info(scanner.Text())
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error streaming logs: %w", err)
	}

	return nil
}

// containerExists checks if the container with the given name exists.
func (l *Logger) containerExists(ctx context.Context, containerName string) (bool, error) {
	output, err := l.runner.RunCommand(ctx, "docker", "ps", "-a", "--format", "{{.Names}}")
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}

	containers := parseOutput(output)
	for _, name := range containers {
		if name == containerName {
			return true, nil
		}
	}
	return false, nil
}

// parseOutput splits the output string into lines.
func parseOutput(output io.Reader) []string {
	scanner := bufio.NewScanner(output)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
