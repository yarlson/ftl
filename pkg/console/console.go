package console

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"golang.org/x/term"
)

var (
	Info = (&pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.InfoMessageStyle,
			Text:  " ",
		},
	}).Println

	Success = (&pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "√",
		},
	}).Println

	Warning = (&pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningMessageStyle,
			Text:  "!",
		},
	}).Println

	ErrPrintln = (&pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorMessageStyle,
			Text:  "✘",
		},
	}).Println

	ErrPrintf = (&pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorMessageStyle,
			Text:  "✘",
		},
	}).Printf

	Input = pterm.FgYellow.Print
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
	spinner := pterm.SpinnerPrinter{
		Sequence:            []string{" ⠋ ", " ⠙ ", " ⠹ ", " ⠸ ", " ⠼ ", " ⠴ ", " ⠦ ", " ⠧ ", " ⠇ ", " ⠏ "},
		Style:               &pterm.ThemeDefault.SpinnerStyle,
		Delay:               time.Millisecond * 200,
		ShowTimer:           false,
		TimerRoundingFactor: time.Second,
		TimerStyle:          &pterm.ThemeDefault.TimerStyle,
		MessageStyle:        pterm.NewStyle(pterm.FgYellow),
		InfoPrinter: &pterm.PrefixPrinter{
			Prefix: pterm.Prefix{
				Style: &pterm.ThemeDefault.InfoMessageStyle,
				Text:  " ",
			},
		},
		SuccessPrinter: &pterm.PrefixPrinter{
			Prefix: pterm.Prefix{
				Style: &pterm.ThemeDefault.SuccessMessageStyle,
				Text:  "√",
			},
		},
		FailPrinter: &pterm.PrefixPrinter{
			Prefix: pterm.Prefix{
				Style: &pterm.ThemeDefault.ErrorMessageStyle,
				Text:  "✘",
			},
		},
		WarningPrinter: &pterm.PrefixPrinter{
			Prefix: pterm.Prefix{
				Style: &pterm.ThemeDefault.WarningMessageStyle,
				Text:  "!",
			},
		},
	}

	sp, _ := spinner.Start(initialText)
	return sp
}
