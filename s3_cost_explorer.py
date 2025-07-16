#!/usr/bin/env python3
"""
AWS S3 Cost Explorer CLI Tool
Retrieve storage costs and storage tiers for S3 buckets
"""

import boto3
import click
import csv
from datetime import datetime, timedelta
from tabulate import tabulate
from botocore.exceptions import ClientError, NoCredentialsError


class S3CostExplorer:
    def __init__(self):
        try:
            # Use default session to pick up AWS CLI configuration
            session = boto3.Session()
            self.s3_client = session.client('s3')
            self.cost_client = session.client('ce', region_name='us-east-1')  # Cost Explorer is only in us-east-1
            self.cloudwatch_client = session.client('cloudwatch')
        except NoCredentialsError:
            click.echo("Error: AWS credentials not found. Please configure AWS CLI or set environment variables.")
            exit(1)

    def get_s3_buckets(self):
        """Get list of all S3 buckets in the account"""
        try:
            response = self.s3_client.list_buckets()
            return [bucket['Name'] for bucket in response['Buckets']]
        except ClientError as e:
            click.echo(f"Error listing buckets: {e}")
            return []

    def get_bucket_storage_tiers(self, bucket_name):
        """Get storage tier breakdown for a specific bucket"""
        try:
            # Get bucket location
            try:
                location_response = self.s3_client.get_bucket_location(Bucket=bucket_name)
                region = location_response['LocationConstraint'] or 'us-east-1'
                if region == 'EU':
                    region = 'eu-west-1'
            except ClientError:
                region = 'us-east-1'
            
            # Create CloudWatch client for the bucket's region using the same session
            session = boto3.Session()
            cw_client = session.client('cloudwatch', region_name=region)
            
            end_date = datetime.now()
            start_date = end_date - timedelta(days=2)
            
            storage_types = [
                'StandardStorage',
                'StandardIAStorage', 
                'ReducedRedundancyStorage',
                'GlacierStorage',
                'DeepArchiveStorage',
                'IntelligentTieringStorage',
                'OneZoneIAStorage'
            ]
            
            tier_data = {}
            
            for storage_type in storage_types:
                try:
                    response = cw_client.get_metric_statistics(
                        Namespace='AWS/S3',
                        MetricName='BucketSizeBytes',
                        Dimensions=[
                            {'Name': 'BucketName', 'Value': bucket_name},
                            {'Name': 'StorageType', 'Value': storage_type}
                        ],
                        StartTime=start_date,
                        EndTime=end_date,
                        Period=86400,
                        Statistics=['Average']
                    )
                    
                    if response['Datapoints']:
                        size_bytes = response['Datapoints'][-1]['Average']
                        if size_bytes > 0:
                            tier_data[storage_type] = size_bytes
                except ClientError:
                    continue
            
            # If no CloudWatch data, try to get basic bucket info
            if not tier_data:
                try:
                    paginator = self.s3_client.get_paginator('list_objects_v2')
                    total_size = 0
                    object_count = 0
                    
                    for page in paginator.paginate(Bucket=bucket_name):
                        if 'Contents' in page:
                            for obj in page['Contents']:
                                total_size += obj['Size']
                                object_count += 1
                    
                    if total_size > 0:
                        tier_data['StandardStorage'] = total_size
                        
                except ClientError:
                    pass
            
            return tier_data
        except ClientError as e:
            click.echo(f"Error getting storage tiers for {bucket_name}: {e}")
            return {}

    def get_s3_costs(self, days=30):
        """Get S3 costs for the specified number of days"""
        try:
            end_date = datetime.now().date()
            start_date = end_date - timedelta(days=days)
            
            response = self.cost_client.get_cost_and_usage(
                TimePeriod={
                    'Start': start_date.strftime('%Y-%m-%d'),
                    'End': end_date.strftime('%Y-%m-%d')
                },
                Granularity='DAILY',
                Metrics=['BlendedCost'],
                GroupBy=[
                    {
                        'Type': 'DIMENSION',
                        'Key': 'SERVICE'
                    }
                ],
                Filter={
                    'Dimensions': {
                        'Key': 'SERVICE',
                        'Values': ['Amazon Simple Storage Service']
                    }
                }
            )
            
            total_cost = 0
            for result in response['ResultsByTime']:
                for group in result['Groups']:
                    if group['Keys'][0] == 'Amazon Simple Storage Service':
                        cost = float(group['Metrics']['BlendedCost']['Amount'])
                        total_cost += cost
            
            return total_cost
        except ClientError as e:
            click.echo(f"Error getting S3 costs: {e}")
            return 0

    def get_detailed_s3_costs(self, days=30):
        """Get detailed S3 costs broken down by usage type"""
        try:
            end_date = datetime.now().date()
            start_date = end_date - timedelta(days=days)
            
            response = self.cost_client.get_cost_and_usage(
                TimePeriod={
                    'Start': start_date.strftime('%Y-%m-%d'),
                    'End': end_date.strftime('%Y-%m-%d')
                },
                Granularity='MONTHLY',
                Metrics=['BlendedCost'],
                GroupBy=[
                    {
                        'Type': 'DIMENSION',
                        'Key': 'USAGE_TYPE'
                    }
                ],
                Filter={
                    'Dimensions': {
                        'Key': 'SERVICE',
                        'Values': ['Amazon Simple Storage Service']
                    }
                }
            )
            
            cost_breakdown = {}
            for result in response['ResultsByTime']:
                for group in result['Groups']:
                    usage_type = group['Keys'][0]
                    cost = float(group['Metrics']['BlendedCost']['Amount'])
                    if cost > 0:
                        cost_breakdown[usage_type] = cost
            
            return cost_breakdown
        except ClientError as e:
            click.echo(f"Error getting detailed S3 costs: {e}")
            return {}

    def format_bytes(self, bytes_value):
        """Format bytes to human readable format"""
        if bytes_value == 0:
            return "0 B"
        
        for unit in ['B', 'KB', 'MB', 'GB', 'TB']:
            if bytes_value < 1024.0:
                return f"{bytes_value:.2f} {unit}"
            bytes_value /= 1024.0
        return f"{bytes_value:.2f} PB"


@click.group()
def cli():
    """AWS S3 Cost Explorer CLI Tool"""
    pass


@cli.command()
@click.option('--days', default=30, help='Number of days to analyze (default: 30)')
@click.option('--csv', 'output_csv', help='Export to CSV file')
def costs(days, output_csv):
    """Get S3 storage costs for the specified period"""
    explorer = S3CostExplorer()
    
    click.echo(f"Retrieving S3 costs for the last {days} days...")
    total_cost = explorer.get_s3_costs(days)
    
    click.echo(f"\nTotal S3 Cost (last {days} days): ${total_cost:.2f}")
    
    detailed_costs = explorer.get_detailed_s3_costs(days)
    if detailed_costs:
        table_data = []
        for usage_type, cost in sorted(detailed_costs.items(), key=lambda x: x[1], reverse=True):
            table_data.append([usage_type, f"${cost:.2f}"])
        
        if output_csv:
            with open(output_csv, 'w', newline='') as csvfile:
                writer = csv.writer(csvfile)
                writer.writerow(['Usage Type', 'Cost'])
                for usage_type, cost in sorted(detailed_costs.items(), key=lambda x: x[1], reverse=True):
                    writer.writerow([usage_type, f"${cost:.2f}"])
            click.echo(f"Cost data exported to {output_csv}")
        else:
            click.echo("\nDetailed Cost Breakdown:")
            click.echo(tabulate(table_data, headers=['Usage Type', 'Cost'], tablefmt='grid'))


@cli.command()
@click.option('--detailed', is_flag=True, help='Show detailed storage tier breakdown')
@click.option('--csv', 'output_csv', help='Export to CSV file')
def buckets(detailed, output_csv):
    """List all S3 buckets with storage information"""
    explorer = S3CostExplorer()
    
    click.echo("Retrieving S3 bucket information...")
    buckets = explorer.get_s3_buckets()
    
    if not buckets:
        click.echo("No S3 buckets found in the account.")
        return
    
    table_data = []
    for bucket in buckets:
        click.echo(f"Analyzing bucket: {bucket}")
        tier_data = explorer.get_bucket_storage_tiers(bucket)
        
        if tier_data:
            if detailed:
                for storage_type, size_bytes in tier_data.items():
                    readable_size = explorer.format_bytes(size_bytes)
                    table_data.append([bucket, storage_type, readable_size])
            else:
                total_size = sum(tier_data.values())
                readable_size = explorer.format_bytes(total_size)
                storage_types = list(tier_data.keys())
                table_data.append([bucket, readable_size, ', '.join(storage_types)])
        else:
            table_data.append([bucket, "No data", "N/A"])
    
    if table_data:
        if output_csv:
            with open(output_csv, 'w', newline='') as csvfile:
                writer = csv.writer(csvfile)
                if detailed:
                    writer.writerow(['Bucket', 'Storage Tier', 'Size'])
                else:
                    writer.writerow(['Bucket', 'Total Size', 'Storage Types'])
                writer.writerows(table_data)
            click.echo(f"Bucket data exported to {output_csv}")
        else:
            if detailed:
                click.echo("\nS3 Bucket Storage Tiers (Detailed):")
                click.echo(tabulate(table_data, headers=['Bucket', 'Storage Tier', 'Size'], tablefmt='grid'))
            else:
                click.echo("\nS3 Bucket Storage Summary:")
                click.echo(tabulate(table_data, headers=['Bucket', 'Total Size', 'Storage Types'], tablefmt='grid'))


@cli.command()
@click.argument('bucket_name')
@click.option('--csv', 'output_csv', help='Export to CSV file')
def bucket_details(bucket_name, output_csv):
    """Get detailed storage tier information for a specific bucket"""
    explorer = S3CostExplorer()
    
    click.echo(f"Retrieving detailed information for bucket: {bucket_name}")
    tier_data = explorer.get_bucket_storage_tiers(bucket_name)
    
    if not tier_data:
        click.echo(f"No storage tier data found for bucket: {bucket_name}")
        return
    
    table_data = []
    total_size = 0
    for storage_type, size_bytes in tier_data.items():
        readable_size = explorer.format_bytes(size_bytes)
        table_data.append([storage_type, readable_size])
        total_size += size_bytes
    
    if output_csv:
        with open(output_csv, 'w', newline='') as csvfile:
            writer = csv.writer(csvfile)
            writer.writerow(['Storage Tier', 'Size'])
            writer.writerows(table_data)
            # Add total size as the last row
            writer.writerow(['Total', explorer.format_bytes(total_size)])
        click.echo(f"Bucket details exported to {output_csv}")
    else:
        click.echo(f"\nStorage Tier Breakdown for {bucket_name}:")
        click.echo(tabulate(table_data, headers=['Storage Tier', 'Size'], tablefmt='grid'))
        click.echo(f"\nTotal Size: {explorer.format_bytes(total_size)}")


if __name__ == '__main__':
    cli()