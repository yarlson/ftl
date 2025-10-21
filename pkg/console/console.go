package console

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Color represents an ANSI color code.
type Color int

// Available colors.
const (
	ColorReset Color = iota
	ColorRed
	ColorGreen
	ColorYellow
)

var disableColor bool

// String returns the ANSI escape code for the given color.
// If colors are disabled, it returns an empty string.
func (c Color) String() string {
	if disableColor {
		return ""
	}

	switch c {
	case ColorReset:
		return "\033[0m"
	case ColorRed:
		return "\033[91m"
	case ColorGreen:
		return "\033[92m"
	case ColorYellow:
		return "\033[93m"
	default:
		return ""
	}
}

func init() {
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		disableColor = true
	}
}

// Info prints an information message.
func Info(a ...interface{}) {
	message := fmt.Sprint(a...)
	fmt.Printf("  %s\n", message)
}

// Success prints a success message.
func Success(a ...interface{}) {
	message := fmt.Sprint(a...)
	fmt.Printf("%s✓%s %s\n", ColorGreen, ColorReset, message)
}

// Warning prints a warning message.
func Warning(a ...interface{}) {
	message := fmt.Sprint(a...)
	fmt.Printf("%s!%s %s\n", ColorYellow, ColorReset, message)
}

// Error prints an error message with a newline.
func Error(a ...interface{}) {
	message := fmt.Sprint(a...)
	fmt.Printf("%s✘%s %s\n", ColorRed, ColorReset, message)
}

// Input prints an input prompt.
func Input(a ...interface{}) {
	message := fmt.Sprint(a...)
	fmt.Printf("%s%s%s", ColorYellow, message, ColorReset)
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
