package explorer

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// GetS3Costs returns the total blended S3 cost over the last `days` days.
func (e *S3CostExplorer) GetS3Costs(ctx context.Context, days int) (float64, error) {
	end := time.Now().UTC().Format("2006-01-02")
	start := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02")

	resp, err := e.CEClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &cetypes.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: cetypes.GranularityDaily,
		Metrics:     []string{"BlendedCost"},
		GroupBy: []cetypes.GroupDefinition{
			{Type: cetypes.GroupDefinitionTypeDimension, Key: aws.String("SERVICE")},
		},
		Filter: &cetypes.Expression{
			Dimensions: &cetypes.DimensionValues{
				Key:    cetypes.DimensionService,
				Values: []string{"Amazon Simple Storage Service"},
			},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("error getting S3 costs: %w", err)
	}

	var total float64
	for _, result := range resp.ResultsByTime {
		for _, group := range result.Groups {
			if len(group.Keys) > 0 && group.Keys[0] == "Amazon Simple Storage Service" {
				amountStr := aws.ToString(group.Metrics["BlendedCost"].Amount)
				cost, parseErr := strconv.ParseFloat(amountStr, 64)
				if parseErr != nil {
					return 0, fmt.Errorf("error parsing cost amount %q: %w", amountStr, parseErr)
				}
				total += cost
			}
		}
	}
	return total, nil
}

// GetDetailedS3Costs returns S3 costs broken down by usage type over the last `days` days.
func (e *S3CostExplorer) GetDetailedS3Costs(ctx context.Context, days int) (map[string]float64, error) {
	end := time.Now().UTC().Format("2006-01-02")
	start := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02")

	resp, err := e.CEClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &cetypes.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: cetypes.GranularityMonthly,
		Metrics:     []string{"BlendedCost"},
		GroupBy: []cetypes.GroupDefinition{
			{Type: cetypes.GroupDefinitionTypeDimension, Key: aws.String("USAGE_TYPE")},
		},
		Filter: &cetypes.Expression{
			Dimensions: &cetypes.DimensionValues{
				Key:    cetypes.DimensionService,
				Values: []string{"Amazon Simple Storage Service"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error getting detailed S3 costs: %w", err)
	}

	breakdown := make(map[string]float64)
	for _, result := range resp.ResultsByTime {
		for _, group := range result.Groups {
			if len(group.Keys) == 0 {
				continue
			}
			usageType := group.Keys[0]
			amountStr := aws.ToString(group.Metrics["BlendedCost"].Amount)
			cost, parseErr := strconv.ParseFloat(amountStr, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("error parsing cost amount %q for %s: %w", amountStr, usageType, parseErr)
			}
			if cost > 0 {
				breakdown[usageType] += cost
			}
		}
	}
	return breakdown, nil
}
