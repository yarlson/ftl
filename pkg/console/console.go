package console

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/pterm/pterm"
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
	Input("Enter server user password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	return string(password), nil
}

// Print prints a message to the console.
func Print(a ...interface{}) {
	fmt.Println(a...)
}
