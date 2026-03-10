package cmd

import (
	"fmt"
	"strconv"

	"github.com/barronlroth/splitwise-cli/internal/api"
	"github.com/barronlroth/splitwise-cli/internal/output"
	"github.com/spf13/cobra"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "List all groups",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.New()
		if err != nil {
			output.Die("%v", err)
		}
		groups, err := client.GetGroups()
		if err != nil {
			output.Die("%v", err)
		}

		if jsonOut {
			output.JSON(groups)
			return
		}

		if quiet {
			for _, g := range groups {
				if g.ID == 0 {
					continue
				}
				fmt.Println(g.ID)
			}
			return
		}

		var rows [][]string
		for _, g := range groups {
			if g.ID == 0 {
				continue // skip non-group expenses
			}
			memberCount := len(g.Members)
			rows = append(rows, []string{
				fmt.Sprintf("%d", g.ID),
				g.Name,
				g.GroupType,
				fmt.Sprintf("%d members", memberCount),
			})
		}
		output.Table([]string{"ID", "Name", "Type", "Members"}, rows)
	},
}

var groupCmd = &cobra.Command{
	Use:   "group <name|id>",
	Short: "Show group details and balances",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.New()
		if err != nil {
			output.Die("%v", err)
		}

		var group *api.Group
		if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			group, err = client.GetGroup(id)
			if err != nil {
				output.Die("%v", err)
			}
		} else {
			group, err = client.ResolveGroupByName(args[0])
			if err != nil {
				output.Die("%v", err)
			}
			// Fetch full details.
			group, err = client.GetGroup(group.ID)
			if err != nil {
				output.Die("%v", err)
			}
		}

		if jsonOut {
			output.JSON(group)
			return
		}

		if quiet {
			fmt.Println(group.ID)
			return
		}

		output.Bold.Printf("%s\n", group.Name)
		fmt.Printf("  ID:   %d\n", group.ID)
		fmt.Printf("  Type: %s\n", group.GroupType)
		fmt.Println()

		// Members and balances.
		output.Bold.Println("Members:")
		for _, m := range group.Members {
			name := m.FirstName + " " + m.LastName
			if len(m.Balance) > 0 {
				for _, b := range m.Balance {
					fmt.Printf("  %-20s %s\n", name, output.FormatAmount(b.Amount, b.CurrencyCode))
				}
			} else {
				fmt.Printf("  %-20s %s\n", name, output.Faint.Sprint("settled up"))
			}
		}

		// Simplified debts.
		if len(group.SimplifiedDebts) > 0 {
			fmt.Println()
			output.Bold.Println("Debts (simplified):")
			memberMap := make(map[int64]string)
			for _, m := range group.Members {
				memberMap[m.ID] = m.FirstName + " " + m.LastName
			}
			for _, d := range group.SimplifiedDebts {
				from := memberMap[d.From]
				to := memberMap[d.To]
				if from == "" {
					from = fmt.Sprintf("User %d", d.From)
				}
				if to == "" {
					to = fmt.Sprintf("User %d", d.To)
				}
				fmt.Printf("  %s → %s: %s %s\n", from, to, d.Amount, d.CurrencyCode)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(groupsCmd)
	rootCmd.AddCommand(groupCmd)
}
