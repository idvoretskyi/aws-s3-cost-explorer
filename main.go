// AWS S3 Cost Explorer CLI Tool
// Retrieve storage costs and storage tiers for S3 buckets.
package main

import (
	"fmt"
	"os"

	"github.com/idvoretskyi/aws-s3-cost-explorer/cmd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aws-s3-cost-explorer",
	Short: "AWS S3 Cost Explorer CLI Tool",
	Long:  "Retrieve storage costs and storage tiers for S3 buckets in your AWS account.",
}

func main() {
	rootCmd.AddCommand(cmd.CostsCmd)
	rootCmd.AddCommand(cmd.BucketsCmd)
	rootCmd.AddCommand(cmd.BucketDetailsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
