package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/paulakimenko/splitwise-cli/internal/api"
	"github.com/paulakimenko/splitwise-cli/internal/config"
	"github.com/paulakimenko/splitwise-cli/internal/output"
	"github.com/spf13/cobra"
)

var expensesCmd = &cobra.Command{
	Use:   "expenses",
	Short: "Manage expenses",
}

var expensesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List expenses",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.New()
		if err != nil {
			output.Die("%v", err)
		}

		groupName, _ := cmd.Flags().GetString("group")
		limit, _ := cmd.Flags().GetInt("limit")
		after, _ := cmd.Flags().GetString("after")
		before, _ := cmd.Flags().GetString("before")

		// Resolve group if specified.
		if groupName == "" {
			cfg, _ := config.Load()
			if cfg != nil {
				groupName = cfg.DefaultGroup
			}
		}

		p := api.GetExpensesParams{
			Limit:       limit,
			DatedAfter:  after,
			DatedBefore: before,
		}

		if groupName != "" {
			group, err := client.ResolveGroupByName(groupName)
			if err != nil {
				output.Die("%v", err)
			}
			p.GroupID = group.ID
		}

		expenses, err := client.GetExpenses(p)
		if err != nil {
			output.Die("%v", err)
		}

		if jsonOut {
			output.JSON(expenses)
			return
		}

		if quiet {
			for _, e := range expenses {
				if e.DeletedAt != nil {
					continue
				}
				fmt.Printf("%d\t%s\t%s\n", e.ID, e.Cost, e.CurrencyCode)
			}
			return
		}

		var rows [][]string
		for _, e := range expenses {
			if e.DeletedAt != nil {
				continue
			}
			date := e.Date
			if t, err := time.Parse(time.RFC3339, e.Date); err == nil {
				date = t.Format("2006-01-02")
			}
			desc := e.Description
			if e.Payment {
				desc = output.Faint.Sprint("💸 " + desc)
			}
			rows = append(rows, []string{
				fmt.Sprintf("%d", e.ID),
				date,
				desc,
				fmt.Sprintf("%s %s", e.Cost, e.CurrencyCode),
			})
		}
		output.Table([]string{"ID", "Date", "Description", "Amount"}, rows)
	},
}

var expensesCreateCmd = &cobra.Command{
	Use:   `create "description" AMOUNT`,
	Short: "Create a new expense",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.New()
		if err != nil {
			output.Die("%v", err)
		}

		description := args[0]
		cost := args[1]
		groupName, _ := cmd.Flags().GetString("group")
		split, _ := cmd.Flags().GetString("split")
		currency, _ := cmd.Flags().GetString("currency")
		paidBy, _ := cmd.Flags().GetString("paid-by")

		// Resolve defaults.
		cfg, _ := config.Load()
		if groupName == "" && cfg != nil {
			groupName = cfg.DefaultGroup
		}
		if currency == "" && cfg != nil {
			currency = cfg.DefaultCurrency
		}

		if groupName == "" {
			output.Die("group is required — pass --group or set default_group")
		}

		group, err := client.ResolveGroupByName(groupName)
		if err != nil {
			output.Die("%v", err)
		}

		p := api.CreateExpenseParams{
			Description:  description,
			Cost:         cost,
			CurrencyCode: currency,
			GroupID:      group.ID,
		}

		if split == "" || split == "even" {
			p.SplitEqually = true
		} else if strings.HasPrefix(split, "exact:") {
			// Custom exact split: "exact:Name:Amount,Name:Amount"
			me, err := client.GetCurrentUser()
			if err != nil {
				output.Die("failed to get current user: %v", err)
			}

			var payerID int64
			if paidBy != "" {
				found := false
				lower := strings.ToLower(paidBy)
				for _, m := range group.Members {
					name := strings.ToLower(m.FirstName + " " + m.LastName)
					if name == lower || strings.ToLower(m.FirstName) == lower {
						payerID = m.ID
						found = true
						break
					}
				}
				if !found {
					output.Die("user not found in group: %s", paidBy)
				}
			} else {
				payerID = me.ID
			}

			// Parse "exact:Name:Amount,Name:Amount"
			pairs := strings.Split(split[6:], ",") // skip "exact:"
			owedMap := make(map[string]string) // lowercase name -> amount
			var owedTotal float64
			for _, pair := range pairs {
				parts := strings.SplitN(pair, ":", 2)
				if len(parts) != 2 {
					output.Die("invalid split format: %s (expected Name:Amount)", pair)
				}
				name := strings.TrimSpace(parts[0])
				amount := strings.TrimSpace(parts[1])
				amtFloat, err := strconv.ParseFloat(amount, 64)
				if err != nil {
					output.Die("invalid amount for %s: %s", name, amount)
				}
				owedMap[strings.ToLower(name)] = fmt.Sprintf("%.2f", amtFloat)
				owedTotal += amtFloat
			}

			// Validate total
			costFloat, err := strconv.ParseFloat(cost, 64)
			if err != nil {
				output.Die("invalid cost: %s", cost)
			}
			if fmt.Sprintf("%.2f", owedTotal) != fmt.Sprintf("%.2f", costFloat) {
				output.Die("split amounts (%.2f) don't add up to total (%.2f)", owedTotal, costFloat)
			}

			// Build shares for each group member
			for _, m := range group.Members {
				paid := "0.00"
				if m.ID == payerID {
					paid = cost
				}
				owed := "0.00"
				lowerFirst := strings.ToLower(m.FirstName)
				lowerFull := strings.ToLower(m.FirstName + " " + m.LastName)
				if amt, ok := owedMap[lowerFirst]; ok {
					owed = amt
				} else if amt, ok := owedMap[lowerFull]; ok {
					owed = amt
				}
				p.Shares = append(p.Shares, api.ShareParam{
					UserID:    m.ID,
					PaidShare: paid,
					OwedShare: owed,
				})
			}
			p.SplitEqually = false
		} else if split == "exact" {
			output.Die("exact split requires amounts — use format: exact:Name:Amount,Name:Amount")
		}

		expense, err := client.CreateExpense(p)
		if err != nil {
			output.Die("%v", err)
		}

		if jsonOut {
			output.JSON(expense)
			return
		}

		if quiet {
			fmt.Println(expense.ID)
			return
		}

		output.Green.Printf("✓ Created expense #%d\n", expense.ID)
		fmt.Printf("  %s — %s %s\n", expense.Description, expense.Cost, expense.CurrencyCode)
	},
}

var expensesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an expense",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.New()
		if err != nil {
			output.Die("%v", err)
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			output.Die("invalid expense ID: %s", args[0])
		}

		if err := client.DeleteExpense(id); err != nil {
			output.Die("%v", err)
		}

		if quiet {
			return
		}

		output.Green.Printf("✓ Deleted expense #%d\n", id)
	},
}

func init() {
	expensesListCmd.Flags().StringP("group", "g", "", "Filter by group name")
	expensesListCmd.Flags().IntP("limit", "l", 20, "Maximum number of expenses")
	expensesListCmd.Flags().String("after", "", "Only expenses after this date (YYYY-MM-DD)")
	expensesListCmd.Flags().String("before", "", "Only expenses before this date (YYYY-MM-DD)")

	expensesCreateCmd.Flags().StringP("group", "g", "", "Group to add expense to")
	expensesCreateCmd.Flags().String("split", "even", `Split type: even, or exact:Name:Amount,Name:Amount (e.g. "exact:MemberA:60,MemberB:40")`)
	expensesCreateCmd.Flags().String("paid-by", "", "Who paid (name, defaults to you)")
	expensesCreateCmd.Flags().StringP("currency", "c", "", "Currency code (e.g. USD)")

	expensesCmd.AddCommand(expensesListCmd)
	expensesCmd.AddCommand(expensesCreateCmd)
	expensesCmd.AddCommand(expensesDeleteCmd)
	rootCmd.AddCommand(expensesCmd)
}
