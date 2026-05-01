// Package explorer provides AWS clients and the core S3CostExplorer type.
package explorer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3CostExplorer holds the AWS service clients used by all commands.
type S3CostExplorer struct {
	S3Client *s3.Client
	CWClient *cloudwatch.Client // default-region; per-bucket clients are created ad-hoc
	CEClient *costexplorer.Client
	Cfg      aws.Config // base config used for per-region client creation
}

// New creates a new S3CostExplorer using the default AWS credential chain.
// It returns an error if credentials or config cannot be loaded.
func New(ctx context.Context) (*S3CostExplorer, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Cost Explorer is only available in us-east-1
	ceCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for Cost Explorer: %w", err)
	}

	return &S3CostExplorer{
		S3Client: s3.NewFromConfig(cfg),
		CWClient: cloudwatch.NewFromConfig(cfg),
		CEClient: costexplorer.NewFromConfig(ceCfg),
		Cfg:      cfg,
	}, nil
}
