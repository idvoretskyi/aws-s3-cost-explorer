package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/idvoretskyi/aws-s3-cost-explorer/internal/explorer"
	"github.com/idvoretskyi/aws-s3-cost-explorer/internal/output"
	"github.com/spf13/cobra"
)

var (
	bucketsDetailed  bool
	bucketsOutputCSV string
)

// BucketsCmd is the `buckets` subcommand.
var BucketsCmd = &cobra.Command{
	Use:   "buckets",
	Short: "List all S3 buckets with storage information",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		e := explorer.New(ctx)

		fmt.Println("Retrieving S3 bucket information...")
		buckets, err := e.GetS3Buckets(ctx)
		if err != nil {
			return err
		}
		if len(buckets) == 0 {
			fmt.Println("No S3 buckets found in the account.")
			return nil
		}

		var headers []string
		if bucketsDetailed {
			headers = []string{"Bucket", "Storage Tier", "Size"}
		} else {
			headers = []string{"Bucket", "Total Size", "Storage Types"}
		}

		var rows [][]string
		for _, bucket := range buckets {
			fmt.Printf("Analyzing bucket: %s\n", bucket)
			tierData, err := e.GetBucketStorageTiers(ctx, bucket)
			if err != nil {
				fmt.Println(err)
			}

			if len(tierData) == 0 {
				rows = append(rows, []string{bucket, "No data", "N/A"})
				continue
			}

			if bucketsDetailed {
				for storageType, sizeBytes := range tierData {
					rows = append(rows, []string{bucket, storageType, explorer.FormatBytes(sizeBytes)})
				}
			} else {
				var totalSize float64
				var types []string
				for k, v := range tierData {
					totalSize += v
					types = append(types, k)
				}
				rows = append(rows, []string{bucket, explorer.FormatBytes(totalSize), strings.Join(types, ", ")})
			}
		}

		if len(rows) == 0 {
			return nil
		}

		if bucketsOutputCSV != "" {
			if err := output.WriteCSV(bucketsOutputCSV, headers, rows); err != nil {
				return err
			}
			fmt.Printf("Bucket data exported to %s\n", bucketsOutputCSV)
		} else {
			if bucketsDetailed {
				fmt.Println("\nS3 Bucket Storage Tiers (Detailed):")
			} else {
				fmt.Println("\nS3 Bucket Storage Summary:")
			}
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

func init() {
	BucketsCmd.Flags().BoolVar(&bucketsDetailed, "detailed", false, "Show detailed storage tier breakdown")
	BucketsCmd.Flags().StringVar(&bucketsOutputCSV, "csv", "", "Export to CSV file")
}
