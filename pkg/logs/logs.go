package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/pterm/pterm"

	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/deployment"
)

var serviceColors = []pterm.Color{
	pterm.FgLightBlue,
	pterm.FgLightGreen,
	pterm.FgLightCyan,
	pterm.FgLightMagenta,
	pterm.FgLightYellow,
	pterm.FgLightRed,
}

// Logger provides methods to fetch logs from remote services.
type Logger struct {
	runner deployment.Runner
}

// NewLogger creates a new Logger instance.
func NewLogger(runner deployment.Runner) *Logger {
	return &Logger{runner: runner}
}

// FetchLogs fetches and optionally streams logs from the specified services.
func (l *Logger) FetchLogs(ctx context.Context, project string, services []string, follow bool) error {
	var wg sync.WaitGroup
	serviceColorMap := make(map[string]pterm.Color)

	// Assign colors to services
	for i, service := range services {
		color := serviceColors[i%len(serviceColors)]
		serviceColorMap[service] = color
	}

	for _, service := range services {
		wg.Add(1)
		go func(svc string) {
			defer wg.Done()

			containerName := svc

			// Check if the container exists
			exists, err := l.containerExists(ctx, containerName)
			if err != nil {
				console.Error(fmt.Sprintf("Failed to check if service %s exists: %v", svc, err))
				return
			}
			if !exists {
				console.Warning(fmt.Sprintf("Service %s is not running on the server", svc))
				return
			}

			// Prepare the command arguments
			cmdArgs := []string{"logs"}
			if follow {
				cmdArgs = append(cmdArgs, "-f")
			}
			cmdArgs = append(cmdArgs, containerName)

			// Run the docker logs command
			reader, err := l.runner.RunCommand(ctx, "docker", cmdArgs...)
			if err != nil {
				console.Error(fmt.Sprintf("Failed to fetch logs for service %s: %v", svc, err))
				return
			}
			defer reader.Close()

			// Prepare the prefix and color
			prefix := fmt.Sprintf("[%s]", svc)
			color := serviceColorMap[svc]
			prefixStyle := pterm.NewStyle(color)
			messageStyle := pterm.NewStyle(pterm.FgDefault)

			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := scanner.Text()
				formattedLine := fmt.Sprintf("%s %s", prefixStyle.Sprint(prefix), messageStyle.Sprint(line))
				console.Print(formattedLine)
			}
			if err := scanner.Err(); err != nil && err != io.EOF {
				console.Error(fmt.Sprintf("Error reading logs for service %s: %v", svc, err))
			}
		}(service)
	}

	wg.Wait()
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
