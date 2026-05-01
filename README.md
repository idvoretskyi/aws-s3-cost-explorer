# AWS S3 Cost Explorer

A CLI tool to retrieve S3 storage costs and storage tiers for buckets in your AWS account.

Written in Go. Produces a single self-contained binary — no runtime or virtualenv required.

## Installation

### From source

```bash
git clone https://github.com/idvoretskyi/aws-s3-cost-explorer.git
cd aws-s3-cost-explorer
go build -o aws-s3-cost-explorer .
```

### Using `go install`

```bash
go install github.com/idvoretskyi/aws-s3-cost-explorer@latest
```

Ensure `$(go env GOPATH)/bin` is on your `$PATH`.

## Prerequisites

AWS credentials must be configured via one of the standard methods:

```bash
aws configure          # AWS CLI profile
# or set environment variables:
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=...
```

## Usage

### Get S3 costs for the last 30 days

```bash
./aws-s3-cost-explorer costs
```

### Get costs for a specific period

```bash
./aws-s3-cost-explorer costs --days 7
```

### Export costs to CSV

```bash
./aws-s3-cost-explorer costs --csv costs.csv
./aws-s3-cost-explorer costs --days 7 --csv costs_7days.csv
```

### List all buckets with storage tier information

```bash
./aws-s3-cost-explorer buckets
```

### List all buckets with detailed storage tier breakdown

```bash
./aws-s3-cost-explorer buckets --detailed
```

### Export bucket information to CSV

```bash
./aws-s3-cost-explorer buckets --csv buckets.csv
./aws-s3-cost-explorer buckets --detailed --csv buckets_detailed.csv
```

### Get detailed storage tier info for a specific bucket

```bash
./aws-s3-cost-explorer bucket-details my-bucket-name
```

### Export bucket details to CSV

```bash
./aws-s3-cost-explorer bucket-details my-bucket-name --csv bucket_details.csv
```

## Features

- Retrieves total S3 costs with detailed breakdown by usage type
- Shows storage tier distribution for each bucket
- Supports all S3 storage classes (Standard, IA, Glacier, Deep Archive, Intelligent-Tiering, One Zone-IA, Reduced Redundancy)
- Falls back to `ListObjectsV2` when CloudWatch metrics are unavailable
- Human-readable size formatting (B / KB / MB / GB / TB / PB)
- Clean grid-style tabular output
- CSV export for all commands

## Required AWS Permissions

| Service | Actions |
|---|---|
| S3 | `s3:ListAllMyBuckets`, `s3:GetBucketLocation`, `s3:ListBucket` |
| CloudWatch | `cloudwatch:GetMetricStatistics` |
| Cost Explorer | `ce:GetCostAndUsage` (requires billing permissions) |
