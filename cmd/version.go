package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set during build time
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of FTL",
	Long:  `Print the version number of FTL deployment tool`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("FTL version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
