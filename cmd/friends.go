package cmd

import (
	"fmt"

	"github.com/paulakimenko/splitwise-cli/internal/api"
	"github.com/paulakimenko/splitwise-cli/internal/output"
	"github.com/spf13/cobra"
)

var friendsCmd = &cobra.Command{
	Use:   "friends",
	Short: "List friends",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.New()
		if err != nil {
			output.Die("%v", err)
		}
		friends, err := client.GetFriends()
		if err != nil {
			output.Die("%v", err)
		}

		if jsonOut {
			output.JSON(friends)
			return
		}

		if quiet {
			for _, f := range friends {
				fmt.Println(f.ID)
			}
			return
		}

		var rows [][]string
		for _, f := range friends {
			name := f.FirstName + " " + f.LastName
			balance := output.Faint.Sprint("settled up")
			if len(f.Balance) > 0 {
				for _, b := range f.Balance {
					balance = output.FormatAmount(b.Amount, b.CurrencyCode)
				}
			}
			rows = append(rows, []string{
				fmt.Sprintf("%d", f.ID),
				name,
				f.Email,
				balance,
			})
		}
		output.Table([]string{"ID", "Name", "Email", "Balance"}, rows)
	},
}

func init() {
	rootCmd.AddCommand(friendsCmd)
}
