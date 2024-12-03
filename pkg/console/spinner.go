package console

import (
	"os"
	"sync"

	"github.com/chelnak/ysmrr"
	"golang.org/x/term"
)

// SpinnerManager handles multiple named spinners with concurrent access support.
type SpinnerManager struct {
	sm       ysmrr.SpinnerManager
	spinners sync.Map
}

// NewSpinnerManager creates a SpinnerManager with the provided writer.
func NewSpinnerManager() *SpinnerManager {
	return &SpinnerManager{
		sm: ysmrr.NewSpinnerManager(),
	}
}

// AddSpinner creates a new spinner with the given name and message.
func (m *SpinnerManager) AddSpinner(name, msg string) *ysmrr.Spinner {
	s := m.sm.AddSpinner(msg)
	m.spinners.Store(name, s)

	return s
}

// UpdateMessage updates the message of the specified spinner.
func (m *SpinnerManager) UpdateMessage(name, msg string) {
	if s, ok := m.spinners.Load(name); ok {
		s.(*ysmrr.Spinner).UpdateMessage(msg)
	}
}

// Complete marks the specified spinner as completed.
func (m *SpinnerManager) Complete(name string) {
	if s, ok := m.spinners.Load(name); ok {
		s.(*ysmrr.Spinner).Complete()
	}
}

// Error marks the specified spinner as errored.
func (m *SpinnerManager) Error(name string) {
	if s, ok := m.spinners.Load(name); ok {
		s.(*ysmrr.Spinner).Error()
	}
}

// ErrorWithMessagef marks the specified spinner as errored with a formatted message.
func (m *SpinnerManager) ErrorWithMessagef(name, format string, args ...interface{}) {
	if s, ok := m.spinners.Load(name); ok {
		s.(*ysmrr.Spinner).ErrorWithMessagef(format, args...)
	}
}

// Start begins the spinner animation for all spinners.
func (m *SpinnerManager) Start() {
	m.sm.Start()
}

func (m *SpinnerManager) Stop() {
	m.sm.Stop()

	if term.IsTerminal(int(os.Stdout.Fd())) {
		print("\033[?25h")
	}
}
