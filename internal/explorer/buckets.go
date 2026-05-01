package explorer

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// GetS3Buckets returns the names of all S3 buckets in the account.
func (e *S3CostExplorer) GetS3Buckets(ctx context.Context) ([]string, error) {
	resp, err := e.S3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("error listing buckets: %w", err)
	}
	names := make([]string, 0, len(resp.Buckets))
	for _, b := range resp.Buckets {
		names = append(names, aws.ToString(b.Name))
	}
	return names, nil
}

// storageTypes mirrors the Python list exactly.
var storageTypes = []string{
	"StandardStorage",
	"StandardIAStorage",
	"ReducedRedundancyStorage",
	"GlacierStorage",
	"DeepArchiveStorage",
	"IntelligentTieringStorage",
	"OneZoneIAStorage",
}

// GetBucketStorageTiers returns a map of storageType -> bytes for the given bucket.
// It queries CloudWatch BucketSizeBytes in the bucket's region, and falls back to
// ListObjectsV2 if no CloudWatch data is available.
func (e *S3CostExplorer) GetBucketStorageTiers(ctx context.Context, bucketName string) (map[string]float64, error) {
	// Determine bucket region
	region := "us-east-1"
	locResp, err := e.S3Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: aws.String(bucketName)})
	if err == nil {
		loc := string(locResp.LocationConstraint)
		if loc == "" {
			region = "us-east-1"
		} else if loc == "EU" {
			region = "eu-west-1"
		} else {
			region = loc
		}
	}

	// Build a per-region CloudWatch client
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("error loading config for region %s: %w", region, err)
	}
	cwClient := cloudwatch.NewFromConfig(cfg)

	endTime := time.Now()
	startTime := endTime.Add(-48 * time.Hour)
	period := int32(86400)

	tierData := make(map[string]float64)

	for _, storageType := range storageTypes {
		resp, err := cwClient.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String("AWS/S3"),
			MetricName: aws.String("BucketSizeBytes"),
			Dimensions: []cwtypes.Dimension{
				{Name: aws.String("BucketName"), Value: aws.String(bucketName)},
				{Name: aws.String("StorageType"), Value: aws.String(storageType)},
			},
			StartTime:  aws.Time(startTime),
			EndTime:    aws.Time(endTime),
			Period:     aws.Int32(period),
			Statistics: []cwtypes.Statistic{cwtypes.StatisticAverage},
		})
		if err != nil {
			continue
		}
		if len(resp.Datapoints) > 0 {
			last := resp.Datapoints[len(resp.Datapoints)-1]
			if aws.ToFloat64(last.Average) > 0 {
				tierData[storageType] = aws.ToFloat64(last.Average)
			}
		}
	}

	// Fallback: list objects and sum sizes
	if len(tierData) == 0 {
		paginator := s3.NewListObjectsV2Paginator(e.S3Client, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		var totalSize float64
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				break
			}
			for _, obj := range page.Contents {
				totalSize += float64(aws.ToInt64(obj.Size))
			}
		}
		if totalSize > 0 {
			tierData["StandardStorage"] = totalSize
		}
	}

	return tierData, nil
}
