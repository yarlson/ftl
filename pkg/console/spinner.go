package console

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/yarlson/ftl/pkg/deployment"
)

// NewSpinner creates a new spinner with the given initial text.
func NewSpinner(initialText string) *pterm.SpinnerPrinter {
	return NewSpinnerWithWriter(initialText, os.Stdout)
}

// NewSpinnerWithWriter creates a new spinner with the given initial text and writer.
func NewSpinnerWithWriter(initialText string, writer io.Writer) *pterm.SpinnerPrinter {
	spinner := pterm.SpinnerPrinter{
		Sequence:            []string{" ⠋ ", " ⠙ ", " ⠹ ", " ⠸ ", " ⠼ ", " ⠴ ", " ⠦ ", " ⠧ ", " ⠇ ", " ⠏ "},
		Style:               &pterm.ThemeDefault.SpinnerStyle,
		Delay:               time.Millisecond * 200,
		ShowTimer:           false,
		TimerRoundingFactor: time.Second,
		TimerStyle:          &pterm.ThemeDefault.TimerStyle,
		MessageStyle:        pterm.NewStyle(pterm.FgYellow),
		InfoPrinter:         infoPrinter,
		SuccessPrinter:      successPrinter,
		FailPrinter:         errorPrinter,
		WarningPrinter:      warningPrinter,
		Writer:              writer,
	}

	spinnerPrinter, _ := spinner.Start(initialText)
	return spinnerPrinter
}

// SpinnerGroup manages spinners for CLI feedback.
type SpinnerGroup struct {
	spinners map[string]*pterm.SpinnerPrinter
	multi    *pterm.MultiPrinter
}

// NewSpinnerGroup creates a new SpinnerGroup.
// If a MultiPrinter is provided, it will use it for spinner output.
func NewSpinnerGroup(multi *pterm.MultiPrinter) *SpinnerGroup {
	return &SpinnerGroup{
		spinners: make(map[string]*pterm.SpinnerPrinter),
		multi:    multi,
	}
}

// RunWithSpinner executes an action while displaying a spinner.
// It handles spinner start, success, and failure messages.
func (sm *SpinnerGroup) RunWithSpinner(message string, action func() error, successMessage string) error {
	var spinner *pterm.SpinnerPrinter
	if sm.multi != nil {
		spinner = NewSpinnerWithWriter(message, sm.multi.NewWriter())
	} else {
		spinner = NewSpinner(message)
	}
	defer func() { _ = spinner.Stop() }()

	if err := action(); err != nil {
		spinner.Fail("Failed")
		return err
	}

	spinner.Success(successMessage)
	return nil
}

// HandleEvent processes deployment events and updates spinners accordingly.
func (sm *SpinnerGroup) HandleEvent(event deployment.Event) error {
	switch event.Type {
	case deployment.EventTypeStart:
		spinner := NewSpinnerWithWriter(event.Message, sm.multi.NewWriter())
		sm.spinners[event.Name] = spinner
	case deployment.EventTypeProgress:
		if spinner, ok := sm.spinners[event.Name]; ok {
			spinner.UpdateText(event.Message)
		} else {
			Info(event.Message)
		}
	case deployment.EventTypeFinish, deployment.EventTypeComplete:
		if spinner, ok := sm.spinners[event.Name]; ok {
			spinner.Success(event.Message)
			delete(sm.spinners, event.Name)
		} else {
			Success(event.Message)
		}
	case deployment.EventTypeError:
		if spinner, ok := sm.spinners[event.Name]; ok {
			spinner.Fail(fmt.Sprintf("Deployment error: %s", event.Message))
			delete(sm.spinners, event.Name)
		} else {
			Error(fmt.Sprintf("Deployment error: %s", event.Message))
		}
		return fmt.Errorf("deployment error: %s", event.Message)
	default:
		if spinner, ok := sm.spinners[event.Name]; ok {
			spinner.UpdateText(event.Message)
		} else {
			Info(event.Message)
		}
	}
	return nil
}

// StopAll stops all active spinners.
func (sm *SpinnerGroup) StopAll() {
	for _, spinner := range sm.spinners {
		_ = spinner.Stop()
	}
}
