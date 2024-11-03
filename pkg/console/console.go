package console

import (
	"bufio"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"golang.org/x/term"
)

var (
	Info       = pterm.Info.Println
	Success    = pterm.Success.Println
	Warning    = pterm.Warning.Println
	ErrPrintln = pterm.Error.Println
	ErrPrintf  = pterm.Error.Printf
	Input      = pterm.FgYellow.Print
)

func ReadLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func ReadPassword() (string, error) {
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	return string(password), nil
}

func NewSpinner(initialText string) *pterm.SpinnerPrinter {
	spinner, _ := pterm.
		DefaultSpinner.
		WithSequence(" ⠋ ", " ⠙ ", " ⠹ ", " ⠸ ", " ⠼ ", " ⠴ ", " ⠦ ", " ⠧ ", " ⠇ ", " ⠏ ").
		WithMessageStyle(pterm.NewStyle(pterm.FgYellow)).
		WithShowTimer(false).
		Start(initialText)

	spinner.SuccessPrinter = &pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "√",
		},
	}

	spinner.InfoPrinter = &pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.InfoMessageStyle,
			Text:  "i",
		},
	}

	spinner.WarningPrinter = &pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningMessageStyle,
			Text:  "!",
		},
	}

	spinner.FailPrinter = &pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorMessageStyle,
			Text:  "✘",
		},
	}

	return spinner
}
