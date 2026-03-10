package cmd

import (
	"fmt"

	"github.com/barronlroth/splitwise-cli/internal/api"
	"github.com/barronlroth/splitwise-cli/internal/output"
	"github.com/spf13/cobra"
)

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show current user info",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.New()
		if err != nil {
			output.Die("%v", err)
		}
		user, err := client.GetCurrentUser()
		if err != nil {
			output.Die("%v", err)
		}

		if jsonOut {
			output.JSON(user)
			return
		}

		if quiet {
			fmt.Println(user.ID)
			return
		}

		output.Bold.Printf("%s %s\n", user.FirstName, user.LastName)
		fmt.Printf("  Email:    %s\n", user.Email)
		fmt.Printf("  ID:       %d\n", user.ID)
		if user.DefaultCurrency != "" {
			fmt.Printf("  Currency: %s\n", user.DefaultCurrency)
		}
		if user.Locale != "" {
			fmt.Printf("  Locale:   %s\n", user.Locale)
		}
	},
}

func init() {
	rootCmd.AddCommand(meCmd)
}
