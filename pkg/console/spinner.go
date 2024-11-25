package console

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// ANSI control codes
const (
	hideCursor   = "\033[?25l"
	showCursor   = "\033[?25h"
	clearLine    = "\r\033[K"
	windowsHide  = "\x1b[?25l"
	windowsShow  = "\x1b[?25h"
	windowsClear = "\r\x1b[K"
)

// EventType represents the type of spinner event
type EventType string

const (
	EventTypeStart    EventType = "start"
	EventTypeProgress EventType = "progress"
	EventTypeFinish   EventType = "finish"
	EventTypeComplete EventType = "complete"
	EventTypeError    EventType = "error"
)

// Event represents a spinner event
type Event struct {
	Type    EventType
	Name    string
	Message string
}

// Spinner represents an animated loading indicator
type Spinner struct {
	chars     []string
	index     int
	lastWrite time.Time
	done      string
	delay     time.Duration
	writer    io.Writer
	message   string
	mutex     sync.Mutex
	stopChan  chan struct{}
	stopped   bool
	isWindows bool
}

// NewSpinner creates a new spinner with default settings
func NewSpinner(initialText string) *Spinner {
	return NewSpinnerWithWriter(initialText, os.Stdout)
}

// NewSpinnerWithWriter creates a new spinner with a custom writer
func NewSpinnerWithWriter(initialText string, writer io.Writer) *Spinner {
	chars := []string{
		"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
	}
	done := "⠿"

	isWindows := runtime.GOOS == "windows"
	if isWindows {
		chars = []string{"-", "\\", "|", "/"}
		done = "-"
	}

	return &Spinner{
		chars:     chars,
		index:     0,
		lastWrite: time.Now(),
		done:      done,
		delay:     200 * time.Millisecond,
		writer:    writer,
		message:   initialText,
		stopChan:  make(chan struct{}),
		isWindows: isWindows,
	}
}

func (s *Spinner) hideCursor() {
	if s.isWindows {
		_, _ = fmt.Fprint(s.writer, windowsHide)
	} else {
		_, _ = fmt.Fprint(s.writer, hideCursor)
	}
}

func (s *Spinner) showCursor() {
	if s.isWindows {
		_, _ = fmt.Fprint(s.writer, windowsShow)
	} else {
		_, _ = fmt.Fprint(s.writer, showCursor)
	}
}

func (s *Spinner) clearLine() {
	if s.isWindows {
		_, _ = fmt.Fprint(s.writer, windowsClear)
	} else {
		_, _ = fmt.Fprint(s.writer, clearLine)
	}
}

// Start begins the spinner animation
func (s *Spinner) Start(text string) (*Spinner, error) {
	s.mutex.Lock()
	if s.stopped {
		s.mutex.Unlock()
		return s, nil
	}
	s.message = text
	s.stopChan = make(chan struct{})
	s.hideCursor()
	s.mutex.Unlock()

	go func() {
		ticker := time.NewTicker(s.delay)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				return
			case <-ticker.C:
				s.mutex.Lock()
				s.clearLine()
				s.write()
				s.mutex.Unlock()
			}
		}
	}()

	return s, nil
}

// Stop halts the spinner animation
func (s *Spinner) Stop() error {
	s.mutex.Lock()
	if !s.stopped {
		s.stopped = true
		close(s.stopChan)
		s.clearLine()
		s.showCursor()
	}
	s.mutex.Unlock()
	return nil
}

// UpdateText changes the spinner message
func (s *Spinner) UpdateText(text string) {
	s.mutex.Lock()
	s.message = text
	s.clearLine()
	s.write()
	s.mutex.Unlock()
}

// Success displays a success message and stops the spinner
func (s *Spinner) Success(text string) {
	s.mutex.Lock()
	if !s.stopped {
		s.stopped = true
		close(s.stopChan)
		s.clearLine()
		_, _ = fmt.Fprintf(s.writer, "%s✔%s %s\n", colorGreen, colorReset, text)
		s.showCursor()
	}
	s.mutex.Unlock()
}

// Fail displays an error message and stops the spinner
func (s *Spinner) Fail(text string) {
	s.mutex.Lock()
	if !s.stopped {
		s.stopped = true
		close(s.stopChan)
		s.clearLine()
		_, _ = fmt.Fprintf(s.writer, "%s✘%s %s\n", colorRed, colorReset, text)
		s.showCursor()
	}
	s.mutex.Unlock()
}

// SpinnerGroup manages multiple spinners
type SpinnerGroup struct {
	spinners map[string]*Spinner
	mutex    sync.Mutex
}

// NewSpinnerGroup creates a new SpinnerGroup
func NewSpinnerGroup() *SpinnerGroup {
	return &SpinnerGroup{
		spinners: make(map[string]*Spinner),
	}
}

// RunWithSpinner executes an action while displaying a spinner
func (sg *SpinnerGroup) RunWithSpinner(message string, action func() error, successMessage string) error {
	spinner := NewSpinner(message)
	_, _ = spinner.Start(message)
	defer func() { _ = spinner.Stop() }()

	if err := action(); err != nil {
		spinner.Fail(fmt.Sprintf("Failed: %s", err))
		return err
	}

	spinner.Success(successMessage)
	return nil
}

// HandleEvent processes spinner events
func (sg *SpinnerGroup) HandleEvent(event Event) error {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	switch event.Type {
	case EventTypeStart:
		spinner := NewSpinner(event.Message)
		_, _ = spinner.Start(event.Message)
		sg.spinners[event.Name] = spinner

	case EventTypeProgress:
		if spinner, ok := sg.spinners[event.Name]; ok {
			spinner.UpdateText(event.Message)
		}

	case EventTypeFinish, EventTypeComplete:
		if spinner, ok := sg.spinners[event.Name]; ok {
			spinner.Success(event.Message)
			delete(sg.spinners, event.Name)
		}

	case EventTypeError:
		if spinner, ok := sg.spinners[event.Name]; ok {
			spinner.Fail(event.Message)
			delete(sg.spinners, event.Name)
		}
		return fmt.Errorf("error: %s", event.Message)
	}

	return nil
}

// StopAll stops all active spinners
func (sg *SpinnerGroup) StopAll() {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	for _, spinner := range sg.spinners {
		_ = spinner.Stop()
	}
}

func (s *Spinner) write() {
	if time.Since(s.lastWrite) > s.delay {
		s.index = (s.index + 1) % len(s.chars)
		s.lastWrite = time.Now()
	}
	_, _ = fmt.Fprintf(s.writer, "%s%s%s %s", colorYellow, s.chars[s.index], colorReset, s.message)
}
