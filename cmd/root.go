package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	jsonOut bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "splitwise",
	Short: "A command-line interface for Splitwise",
	Long:  "Manage your Splitwise groups, expenses, and balances from the terminal.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		noColor, _ := cmd.Flags().GetBool("no-color")
		if noColor || os.Getenv("NO_COLOR") != "" {
			color.NoColor = true
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Minimal output for scripting")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")

	rootCmd.Version = version
	rootCmd.SetVersionTemplate("splitwise {{.Version}}\n")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
