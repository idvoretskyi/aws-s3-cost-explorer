# AWS S3 Cost Explorer

A simple CLI tool to retrieve S3 storage costs and storage tiers for buckets in your AWS account.

## Installation

1. Create a virtual environment and install dependencies:
```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

2. Ensure AWS credentials are configured:
```bash
aws configure
```

## Usage

Make sure to activate the virtual environment before running commands:
```bash
source venv/bin/activate
```

### Get S3 costs for the last 30 days:
```bash
python s3_cost_explorer.py costs
```

### Get costs for a specific period:
```bash
python s3_cost_explorer.py costs --days 7
```

### List all buckets with storage tier information:
```bash
python s3_cost_explorer.py buckets
```

### List all buckets with detailed storage tier breakdown:
```bash
python s3_cost_explorer.py buckets --detailed
```

### Get detailed storage tier info for a specific bucket:
```bash
python s3_cost_explorer.py bucket-details my-bucket-name
```

## Features

- Retrieves total S3 costs with detailed breakdown by usage type
- Shows storage tier distribution for each bucket
- Supports multiple storage classes (Standard, IA, Glacier, Deep Archive, etc.)
- Human-readable size formatting
- Clean tabular output

## Requirements

- Python 3.6+
- AWS CLI configured with appropriate permissions
- Cost Explorer API access (may require billing permissions)
- CloudWatch metrics access for storage tier analysis