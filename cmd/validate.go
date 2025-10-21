package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yarlson/pin"

	"github.com/yarlson/ftl/pkg/console"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your ftl.yaml configuration",
	Long: `Validate checks your ftl.yaml configuration file for errors.
This command performs validation checks including:
- Required fields presence
- Port number validity
- File path existence
- Environment variable resolution
- Service name uniqueness
- Volume reference validity`,
	Run: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) {
	pValidate := pin.New("Validating configuration", pin.WithSpinnerColor(pin.ColorCyan))

	cancelValidate := pValidate.Start(context.Background())
	defer cancelValidate()

	_, err := parseConfig("ftl.yaml")
	if err != nil {
		pValidate.Fail(fmt.Sprintf("Configuration validation failed: %v", err))
		return
	}

	// Additional validation checks could be added here

	pValidate.Stop("Configuration is valid")
	console.Success("All validation checks passed successfully")
}
