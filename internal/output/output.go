package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

var (
	Green  = color.New(color.FgGreen)
	Red    = color.New(color.FgRed)
	Yellow = color.New(color.FgYellow)
	Cyan   = color.New(color.FgCyan)
	Bold   = color.New(color.Bold)
	Faint  = color.New(color.Faint)
)

// JSON prints any value as formatted JSON.
func JSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// Table prints a nicely formatted table to stdout.
func Table(headers []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetRowSeparator("")
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows)
	table.Render()
}

// FormatAmount colorizes a balance amount string.
func FormatAmount(amount string, currencyCode string) string {
	if amount == "" || amount == "0" || amount == "0.0" || amount == "0.00" {
		return Faint.Sprint("settled up")
	}
	val := amount
	if strings.HasPrefix(val, "-") {
		return Red.Sprintf("%s %s", val, currencyCode)
	}
	return Green.Sprintf("%s %s", val, currencyCode)
}

// Die prints an error and exits.
func Die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
