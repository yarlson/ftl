package logs

import (
	"bufio"
	"container/heap"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

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

// LogEntry represents a single log line with its timestamp and service info.
type LogEntry struct {
	Timestamp time.Time
	Line      string
	Service   string
	Color     pterm.Color
}

// LogEntryHeap is a min-heap of LogEntry based on Timestamp
type LogEntryHeap []LogEntry

func (h LogEntryHeap) Len() int { return len(h) }
func (h LogEntryHeap) Less(i, j int) bool {
	return h[i].Timestamp.Before(h[j].Timestamp)
}
func (h LogEntryHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *LogEntryHeap) Push(x interface{}) {
	*h = append(*h, x.(LogEntry))
}

func (h *LogEntryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// FetchLogs fetches and optionally streams logs from the specified services.
func (l *Logger) FetchLogs(ctx context.Context, project string, services []string, follow bool, tail int) error {
	if follow {
		return l.streamLogs(ctx, project, services, tail)
	} else {
		return l.fetchAndSortLogs(ctx, project, services, tail)
	}
}

// fetchAndSortLogs fetches logs from services, sorts them by timestamp, and prints them.
func (l *Logger) fetchAndSortLogs(ctx context.Context, project string, services []string, tail int) error {
	var wg sync.WaitGroup
	logEntries := make([]LogEntry, 0)
	var mu sync.Mutex

	serviceColorMap := assignColorsToServices(services)

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
			cmdArgs := []string{"logs", "--timestamps"}
			if tail >= 0 {
				cmdArgs = append(cmdArgs, fmt.Sprintf("--tail=%d", tail))
			}
			cmdArgs = append(cmdArgs, containerName)

			// Run the docker logs command
			reader, err := l.runner.RunCommand(ctx, "docker", cmdArgs...)
			if err != nil {
				console.Error(fmt.Sprintf("Failed to fetch logs for service %s: %v", svc, err))
				return
			}
			defer reader.Close()

			color := serviceColorMap[svc]

			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := scanner.Text()
				entry, err := parseLogLine(line, svc, color)
				if err != nil {
					// Ignore lines that cannot be parsed
					continue
				}
				mu.Lock()
				logEntries = append(logEntries, entry)
				mu.Unlock()
			}
			if err := scanner.Err(); err != nil && err != io.EOF {
				console.Error(fmt.Sprintf("Error reading logs for service %s: %v", svc, err))
			}
		}(service)
	}

	wg.Wait()

	// Sort the log entries by timestamp
	sort.Slice(logEntries, func(i, j int) bool {
		return logEntries[i].Timestamp.Before(logEntries[j].Timestamp)
	})

	// Print the sorted log entries
	for _, entry := range logEntries {
		prefixStyle := pterm.NewStyle(entry.Color)
		messageStyle := pterm.NewStyle(pterm.FgDefault)
		prefix := fmt.Sprintf("[%s]", entry.Service)
		formattedLine := fmt.Sprintf("%s %s", prefixStyle.Sprint(prefix), messageStyle.Sprint(entry.Line))
		console.Print(formattedLine)
	}

	return nil
}

// streamLogs streams logs from services, merging them in real-time by timestamp.
func (l *Logger) streamLogs(ctx context.Context, project string, services []string, tail int) error {
	serviceColorMap := assignColorsToServices(services)

	type logStream struct {
		entries chan LogEntry
		done    chan struct{}
	}

	streams := make(map[string]*logStream)
	var wg sync.WaitGroup

	// Start a goroutine for each service to read logs
	for _, service := range services {
		entries := make(chan LogEntry, 100)
		done := make(chan struct{})
		streams[service] = &logStream{entries: entries, done: done}

		wg.Add(1)
		go func(svc string) {
			defer wg.Done()
			defer close(entries)
			defer close(done)

			// Check if the container exists
			exists, err := l.containerExists(ctx, svc)
			if err != nil {
				console.Error(fmt.Sprintf("Failed to check if service %s exists: %v", svc, err))
				return
			}
			if !exists {
				console.Warning(fmt.Sprintf("Service %s is not running on the server", svc))
				return
			}

			// Prepare the command arguments
			cmdArgs := []string{"logs", "--timestamps"}
			if tail >= 0 {
				cmdArgs = append(cmdArgs, fmt.Sprintf("--tail=%d", tail))
			}
			cmdArgs = append(cmdArgs, "-f", svc)

			// Run the docker logs command
			reader, err := l.runner.RunCommand(ctx, "docker", cmdArgs...)
			if err != nil {
				console.Error(fmt.Sprintf("Failed to fetch logs for service %s: %v", svc, err))
				return
			}
			defer reader.Close()

			color := serviceColorMap[svc]

			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := scanner.Text()
				entry, err := parseLogLine(line, svc, color)
				if err != nil {
					// Ignore lines that cannot be parsed
					continue
				}
				select {
				case entries <- entry:
				case <-ctx.Done():
					return
				}
			}
			if err := scanner.Err(); err != nil && err != io.EOF {
				console.Error(fmt.Sprintf("Error reading logs for service %s: %v", svc, err))
			}
		}(service)
	}

	// Merge logs from all services
	h := &LogEntryHeap{}
	heap.Init(h)

	for {
		// Fill the heap with at least one log entry from each service
		for svc, stream := range streams {
			select {
			case entry, ok := <-stream.entries:
				if ok {
					heap.Push(h, entry)
				} else {
					// The service has no more logs
					delete(streams, svc)
				}
			default:
				// No new entries at the moment
			}
		}

		if h.Len() == 0 {
			if len(streams) == 0 {
				// All streams are done
				break
			}
			// Sleep briefly to wait for new log entries
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Pop the earliest log entry and print it
		entry := heap.Pop(h).(LogEntry)
		prefixStyle := pterm.NewStyle(entry.Color)
		messageStyle := pterm.NewStyle(pterm.FgDefault)
		prefix := fmt.Sprintf("[%s]", entry.Service)
		formattedLine := fmt.Sprintf("%s %s", prefixStyle.Sprint(prefix), messageStyle.Sprint(entry.Line))
		console.Print(formattedLine)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	return nil
}

// parseLogLine parses a log line with a timestamp
func parseLogLine(line, service string, color pterm.Color) (LogEntry, error) {
	// Expected format: "<timestamp> <log message>"
	idx := strings.Index(line, " ")
	if idx == -1 {
		return LogEntry{}, fmt.Errorf("invalid log line format")
	}
	timestampStr := line[:idx]
	message := line[idx+1:]

	timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid timestamp format")
	}

	return LogEntry{
		Timestamp: timestamp,
		Line:      message,
		Service:   service,
		Color:     color,
	}, nil
}

// assignColorsToServices assigns colors to services
func assignColorsToServices(services []string) map[string]pterm.Color {
	serviceColorMap := make(map[string]pterm.Color)
	for i, service := range services {
		color := serviceColors[i%len(serviceColors)]
		serviceColorMap[service] = color
	}
	return serviceColorMap
}

// containerExists checks if the container with the given name exists.
func (l *Logger) containerExists(ctx context.Context, containerName string) (bool, error) {
	outputReader, err := l.runner.RunCommand(ctx, "docker", "ps", "-a", "--format", "{{.Names}}")
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}
	defer outputReader.Close()

	containers := parseOutput(outputReader)
	for _, name := range containers {
		if name == containerName {
			return true, nil
		}
	}
	return false, nil
}

// parseOutput reads lines from the output reader.
func parseOutput(output io.Reader) []string {
	scanner := bufio.NewScanner(output)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
