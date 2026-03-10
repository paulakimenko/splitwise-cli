package cmd

import (
	"fmt"

	"github.com/barronlroth/splitwise-cli/internal/api"
	"github.com/barronlroth/splitwise-cli/internal/config"
	"github.com/barronlroth/splitwise-cli/internal/output"
	"github.com/spf13/cobra"
)

var balancesCmd = &cobra.Command{
	Use:   "balances",
	Short: "Show who owes whom",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.New()
		if err != nil {
			output.Die("%v", err)
		}

		groupName, _ := cmd.Flags().GetString("group")
		if groupName == "" {
			cfg, _ := config.Load()
			if cfg != nil {
				groupName = cfg.DefaultGroup
			}
		}

		if groupName != "" {
			showGroupBalances(client, groupName)
		} else {
			showFriendBalances(client)
		}
	},
}

func showGroupBalances(client *api.Client, groupName string) {
	group, err := client.ResolveGroupByName(groupName)
	if err != nil {
		output.Die("%v", err)
	}
	group, err = client.GetGroup(group.ID)
	if err != nil {
		output.Die("%v", err)
	}

	if jsonOut {
		output.JSON(group.SimplifiedDebts)
		return
	}

	memberMap := make(map[int64]string)
	for _, m := range group.Members {
		memberMap[m.ID] = m.FirstName + " " + m.LastName
	}

	if len(group.SimplifiedDebts) == 0 {
		fmt.Println("All settled up! 🎉")
		return
	}

	output.Bold.Printf("Balances for %s:\n\n", group.Name)
	var rows [][]string
	for _, d := range group.SimplifiedDebts {
		from := memberMap[d.From]
		to := memberMap[d.To]
		if from == "" {
			from = fmt.Sprintf("User %d", d.From)
		}
		if to == "" {
			to = fmt.Sprintf("User %d", d.To)
		}
		rows = append(rows, []string{
			from,
			"→",
			to,
			output.Red.Sprintf("%s %s", d.Amount, d.CurrencyCode),
		})
	}
	output.Table([]string{"From", "", "To", "Amount"}, rows)
}

func showFriendBalances(client *api.Client) {
	friends, err := client.GetFriends()
	if err != nil {
		output.Die("%v", err)
	}

	if jsonOut {
		output.JSON(friends)
		return
	}

	hasBalance := false
	var rows [][]string
	for _, f := range friends {
		for _, b := range f.Balance {
			if b.Amount == "0" || b.Amount == "0.0" || b.Amount == "0.00" {
				continue
			}
			hasBalance = true
			name := f.FirstName + " " + f.LastName
			rows = append(rows, []string{
				name,
				output.FormatAmount(b.Amount, b.CurrencyCode),
			})
		}
	}

	if !hasBalance {
		fmt.Println("All settled up! 🎉")
		return
	}

	output.Table([]string{"Friend", "Balance"}, rows)
}

func init() {
	balancesCmd.Flags().StringP("group", "g", "", "Show balances for a specific group")
	rootCmd.AddCommand(balancesCmd)
}
