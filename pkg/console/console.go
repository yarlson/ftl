package console

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var (
	colorReset  = "\033[0m"
	colorRed    = "\033[91m" // Bright Red
	colorGreen  = "\033[92m" // Bright Green
	colorYellow = "\033[93m" // Bright Yellow
)

func init() {
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		colorReset = ""
		colorRed = ""
		colorGreen = ""
		colorYellow = ""
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
	fmt.Printf("%s✓%s %s\n", colorGreen, colorReset, message)
}

// Warning prints a warning message.
func Warning(a ...interface{}) {
	message := fmt.Sprint(a...)
	fmt.Printf("%s!%s %s\n", colorYellow, colorReset, message)
}

// Error prints an error message with a newline.
func Error(a ...interface{}) {
	message := fmt.Sprint(a...)
	fmt.Printf("%s✘%s %s\n", colorRed, colorReset, message)
}

// Input prints an input prompt.
func Input(a ...interface{}) {
	message := fmt.Sprint(a...)
	fmt.Printf("%s%s%s", colorYellow, message, colorReset)
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

// Reset ensures the cursor is visible and terminal is in a normal state.
func Reset() {
	_ = os.Stdout.Sync()
	fmt.Print("\033[?25h")
}

func ClearPreviousLine() {
	fmt.Print("\033[1A\033[K")
}
