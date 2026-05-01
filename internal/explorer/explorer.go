// Package explorer provides AWS clients and the core S3CostExplorer type.
package explorer

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3CostExplorer holds the AWS service clients used by all commands.
type S3CostExplorer struct {
	S3Client   *s3.Client
	CWClient   *cloudwatch.Client // default-region; per-bucket clients are created ad-hoc
	CEClient   *costexplorer.Client
	DefaultCfg interface{} // aws.Config stored as interface to allow region override
}

// New creates a new S3CostExplorer using the default AWS credential chain.
// It exits with a message if credentials cannot be loaded.
func New(ctx context.Context) *S3CostExplorer {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: AWS credentials not found. Please configure AWS CLI or set environment variables.\n")
		os.Exit(1)
	}

	// Cost Explorer is only available in us-east-1
	ceCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading AWS config for Cost Explorer: %v\n", err)
		os.Exit(1)
	}

	return &S3CostExplorer{
		S3Client: s3.NewFromConfig(cfg),
		CWClient: cloudwatch.NewFromConfig(cfg),
		CEClient: costexplorer.NewFromConfig(ceCfg),
	}
}
