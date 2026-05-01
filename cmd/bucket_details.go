package cmd

import (
	"context"
	"fmt"

	"github.com/idvoretskyi/aws-s3-cost-explorer/internal/explorer"
	"github.com/idvoretskyi/aws-s3-cost-explorer/internal/output"
	"github.com/spf13/cobra"
)

var bucketDetailsOutputCSV string

// BucketDetailsCmd is the `bucket-details` subcommand.
var BucketDetailsCmd = &cobra.Command{
	Use:   "bucket-details <bucket-name>",
	Short: "Get detailed storage tier information for a specific bucket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]
		ctx := context.Background()
		e := explorer.New(ctx)

		fmt.Printf("Retrieving detailed information for bucket: %s\n", bucketName)
		tierData, err := e.GetBucketStorageTiers(ctx, bucketName)
		if err != nil {
			return err
		}
		if len(tierData) == 0 {
			fmt.Printf("No storage tier data found for bucket: %s\n", bucketName)
			return nil
		}

		headers := []string{"Storage Tier", "Size"}
		var rows [][]string
		var totalSize float64
		for storageType, sizeBytes := range tierData {
			rows = append(rows, []string{storageType, explorer.FormatBytes(sizeBytes)})
			totalSize += sizeBytes
		}

		if bucketDetailsOutputCSV != "" {
			// Append a Total row to the CSV (matches Python behaviour)
			csvRows := append(rows, []string{"Total", explorer.FormatBytes(totalSize)})
			if err := output.WriteCSV(bucketDetailsOutputCSV, headers, csvRows); err != nil {
				return err
			}
			fmt.Printf("Bucket details exported to %s\n", bucketDetailsOutputCSV)
		} else {
			fmt.Printf("\nStorage Tier Breakdown for %s:\n", bucketName)
			output.PrintTable(headers, rows)
			fmt.Printf("\nTotal Size: %s\n", explorer.FormatBytes(totalSize))
		}
		return nil
	},
}

func init() {
	BucketDetailsCmd.Flags().StringVar(&bucketDetailsOutputCSV, "csv", "", "Export to CSV file")
}
