package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/idvoretskyi/aws-s3-cost-explorer/internal/explorer"
	"github.com/idvoretskyi/aws-s3-cost-explorer/internal/output"
	"github.com/spf13/cobra"
)

var (
	costsDays      int
	costsOutputCSV string
)

// CostsCmd is the `costs` subcommand.
var CostsCmd = &cobra.Command{
	Use:   "costs",
	Short: "Get S3 storage costs for the specified period",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		e, err := explorer.New(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("Retrieving S3 costs for the last %d days...\n", costsDays)
		total, err := e.GetS3Costs(ctx, costsDays)
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal S3 Cost (last %d days): $%.2f\n", costsDays, total)

		detailed, err := e.GetDetailedS3Costs(ctx, costsDays)
		if err != nil {
			return err
		}
		if len(detailed) == 0 {
			return nil
		}

		// Sort by cost descending
		type kv struct {
			Key   string
			Value float64
		}
		var sorted []kv
		for k, v := range detailed {
			sorted = append(sorted, kv{k, v})
		}
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

		headers := []string{"Usage Type", "Cost"}
		var rows [][]string
		for _, item := range sorted {
			rows = append(rows, []string{item.Key, fmt.Sprintf("$%.2f", item.Value)})
		}

		if costsOutputCSV != "" {
			if err := output.WriteCSV(costsOutputCSV, headers, rows); err != nil {
				return err
			}
			fmt.Printf("Cost data exported to %s\n", costsOutputCSV)
		} else {
			fmt.Println("\nDetailed Cost Breakdown:")
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

func init() {
	CostsCmd.Flags().IntVar(&costsDays, "days", 30, "Number of days to analyze (default: 30)")
	CostsCmd.Flags().StringVar(&costsOutputCSV, "csv", "", "Export to CSV file")
}
