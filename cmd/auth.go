package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/barronlroth/splitwise-cli/internal/auth"
	"github.com/barronlroth/splitwise-cli/internal/output"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Splitwise via OAuth 2.0",
	Long: `Authenticate with Splitwise using OAuth 2.0.

You'll need a Splitwise API key. Register your app at:
  https://secure.splitwise.com/apps

Set the callback URL to http://localhost (any port).
You'll be prompted for your Client ID and Client Secret.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Splitwise OAuth Setup")
		fmt.Println("Register an app at https://secure.splitwise.com/apps")
		fmt.Println()

		fmt.Print("Client ID: ")
		clientID, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)
		if clientID == "" {
			output.Die("client ID is required")
		}

		fmt.Print("Client Secret: ")
		clientSecret, _ := reader.ReadString('\n')
		clientSecret = strings.TrimSpace(clientSecret)
		if clientSecret == "" {
			output.Die("client secret is required")
		}

		if err := auth.Login(clientID, clientSecret); err != nil {
			output.Die("%v", err)
		}

		output.Green.Println("✓ Authenticated successfully!")
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Run: func(cmd *cobra.Command, args []string) {
		if err := auth.Logout(); err != nil {
			output.Die("%v", err)
		}
		fmt.Println("Logged out.")
	},
}

func init() {
	authCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(authCmd)
}
