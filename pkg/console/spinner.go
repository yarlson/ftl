package console

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// Event represents the type of spinner event.
type Event struct {
	Type    EventType
	Name    string
	Message string
}

// EventType defines the possible states of a spinner.
type EventType string

const (
	EventStart    EventType = "start"
	EventProgress EventType = "progress"
	EventFinish   EventType = "finish"
	EventComplete EventType = "complete"
	EventError    EventType = "error"
)

const (
	defaultDelay = 200 * time.Millisecond
)

// terminal control sequences
const (
	escHideCursor = "\033[?25l"
	escShowCursor = "\033[?25h"
	escClearLine  = "\r\033[K"
	// Windows-specific sequences
	winHideCursor = "\x1b[?25l"
	winShowCursor = "\x1b[?25h"
	winClearLine  = "\r\x1b[K"
)

// Spinner represents an animated loading indicator.
type Spinner struct {
	frames    []string
	done      string
	curr      int
	lastWrite time.Time
	delay     time.Duration
	w         io.Writer
	msg       string
	mu        sync.Mutex
	stop      chan struct{}
	stopped   bool
	isWin     bool
}

// New creates a new spinner with default settings writing to os.Stdout.
func New(msg string) *Spinner {
	return NewWithWriter(msg, os.Stdout)
}

// NewWithWriter creates a new spinner that writes to the provided writer.
func NewWithWriter(msg string, w io.Writer) *Spinner {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := "⠿"

	isWin := runtime.GOOS == "windows"
	if isWin {
		frames = []string{"-", "\\", "|", "/"}
		done = "-"
	}

	return &Spinner{
		frames: frames,
		done:   done,
		delay:  defaultDelay,
		w:      w,
		msg:    msg,
		stop:   make(chan struct{}),
		isWin:  isWin,
	}
}

// Start begins the spinner animation.
func (s *Spinner) Start(msg string) (*Spinner, error) {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return s, nil
	}
	s.msg = msg
	s.stop = make(chan struct{})
	s.hideCursor()
	s.mu.Unlock()

	go s.run()
	return s, nil
}

func (s *Spinner) run() {
	ticker := time.NewTicker(s.delay)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.mu.Lock()
			s.clearLine()
			s.write()
			s.mu.Unlock()
		}
	}
}

// Stop halts the spinner animation.
func (s *Spinner) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return nil
	}

	s.stopped = true
	close(s.stop)
	s.clearLine()
	s.showCursor()
	return nil
}

// UpdateMessage changes the spinner message.
func (s *Spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.msg = msg
	s.clearLine()
	s.write()
}

// Success displays a success message and stops the spinner.
func (s *Spinner) Success(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return
	}

	s.stopped = true
	close(s.stop)
	s.clearLine()
	_, _ = fmt.Fprintf(s.w, "%s✔%s %s\n", colorGreen, colorReset, msg)
	s.showCursor()
}

// Fail displays an error message and stops the spinner.
func (s *Spinner) Fail(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return
	}

	s.stopped = true
	close(s.stop)
	s.clearLine()
	_, _ = fmt.Fprintf(s.w, "%s✘%s %s\n", colorRed, colorReset, msg)
	s.showCursor()
}

// Group manages multiple concurrent spinners.
type Group struct {
	spinners map[string]*Spinner
	mu       sync.Mutex
}

// NewGroup creates a new spinner group.
func NewGroup() *Group {
	return &Group{
		spinners: make(map[string]*Spinner),
	}
}

// RunWithSpinner executes an action while displaying a spinner.
func (g *Group) RunWithSpinner(msg string, action func() error, successMsg string) error {
	s := New(msg)
	if _, err := s.Start(msg); err != nil {
		return fmt.Errorf("failed to start spinner: %w", err)
	}
	defer func() { _ = s.Stop() }()

	if err := action(); err != nil {
		s.Fail(fmt.Sprintf("Failed: %v", err))
		return err
	}

	s.Success(successMsg)
	return nil
}

// HandleEvent processes spinner events.
func (g *Group) HandleEvent(e Event) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	switch e.Type {
	case EventStart:
		s := New(e.Message)
		if _, err := s.Start(e.Message); err != nil {
			return fmt.Errorf("failed to start spinner: %w", err)
		}
		g.spinners[e.Name] = s

	case EventProgress:
		if s, ok := g.spinners[e.Name]; ok {
			s.UpdateMessage(e.Message)
		}

	case EventFinish, EventComplete:
		if s, ok := g.spinners[e.Name]; ok {
			s.Success(e.Message)
			delete(g.spinners, e.Name)
		}

	case EventError:
		if s, ok := g.spinners[e.Name]; ok {
			s.Fail(e.Message)
			delete(g.spinners, e.Name)
			return fmt.Errorf("error: %s", e.Message)
		}
	}

	return nil
}

// StopAll stops all active spinners in the group.
func (g *Group) StopAll() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, s := range g.spinners {
		_ = s.Stop()
	}
}

// Terminal control methods
func (s *Spinner) hideCursor() {
	if s.isWin {
		_, _ = fmt.Fprint(s.w, winHideCursor)
		return
	}
	_, _ = fmt.Fprint(s.w, escHideCursor)
}

func (s *Spinner) showCursor() {
	if s.isWin {
		_, _ = fmt.Fprint(s.w, winShowCursor)
		return
	}
	_, _ = fmt.Fprint(s.w, escShowCursor)
}

func (s *Spinner) clearLine() {
	if s.isWin {
		_, _ = fmt.Fprint(s.w, winClearLine)
		return
	}
	_, _ = fmt.Fprint(s.w, escClearLine)
}

func (s *Spinner) write() {
	if time.Since(s.lastWrite) > s.delay {
		s.curr = (s.curr + 1) % len(s.frames)
		s.lastWrite = time.Now()
	}
	_, _ = fmt.Fprintf(s.w, "%s%s%s %s", colorYellow, s.frames[s.curr], colorReset, s.msg)
}
