package cmd

import (
	"fmt"

	"github.com/barronlroth/splitwise-cli/internal/config"
	"github.com/barronlroth/splitwise-cli/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  default_group      Default group name for commands
  default_currency   Default currency code (e.g. USD, EUR)`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Set(args[0], args[1]); err != nil {
			output.Die("%v", err)
		}
		fmt.Printf("Set %s = %s\n", args[0], args[1])
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			output.Die("%v", err)
		}

		if jsonOut {
			output.JSON(cfg)
			return
		}

		fmt.Printf("Config file: %s\n\n", config.ConfigPath())
		if cfg.DefaultGroup != "" {
			fmt.Printf("  default_group:    %s\n", cfg.DefaultGroup)
		}
		if cfg.DefaultCurrency != "" {
			fmt.Printf("  default_currency: %s\n", cfg.DefaultCurrency)
		}
		if cfg.DefaultGroup == "" && cfg.DefaultCurrency == "" {
			fmt.Println("  (no values set)")
		}
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}
