package console

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"golang.org/x/term"
)

// Info prints an information message.
func Info(a ...interface{}) {
	infoPrinter.Println(a...)
}

var infoPrinter = &pterm.PrefixPrinter{
	Prefix: pterm.Prefix{
		Style: &pterm.ThemeDefault.InfoMessageStyle,
		Text:  " ",
	},
}

// Success prints a success message.
func Success(a ...interface{}) {
	successPrinter.Println(a...)
}

var successPrinter = &pterm.PrefixPrinter{
	Prefix: pterm.Prefix{
		Style: &pterm.ThemeDefault.SuccessMessageStyle,
		Text:  "√",
	},
}

// Warning prints a warning message.
func Warning(a ...interface{}) {
	warningPrinter.Println(a...)
}

var warningPrinter = &pterm.PrefixPrinter{
	Prefix: pterm.Prefix{
		Style: &pterm.ThemeDefault.WarningMessageStyle,
		Text:  "!",
	},
}

// Error prints an error message with a newline.
func Error(a ...interface{}) {
	errorPrinter.Println(a...)
}

var errorPrinter = &pterm.PrefixPrinter{
	Prefix: pterm.Prefix{
		Style: &pterm.ThemeDefault.ErrorMessageStyle,
		Text:  "✘",
	},
}

// Input prints an input prompt.
func Input(a ...interface{}) {
	pterm.FgYellow.Print(a...)
}

// ReadLine reads a line from standard input.
func ReadLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// ReadPassword reads a password from standard input without echoing.
func ReadPassword() (string, error) {
	Input("Password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	Input("\n") // Move to the next line after input
	return string(password), nil
}

func NewSpinner(initialText string) *pterm.SpinnerPrinter {
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
	}

	spinnerPrinter, _ := spinner.Start(initialText)
	return spinnerPrinter
}
