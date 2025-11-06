# Terraform Infrastructure for PubMed API

This directory contains Terraform configuration for provisioning AWS infrastructure for the PubMed API.

## Overview

The Terraform configuration creates:
- **ECR Repository**: For storing Docker images
- **S3 Bucket**: For storing article data (JSONL files)

## Prerequisites

- Terraform >= 1.0
- AWS CLI configured with appropriate credentials
- AWS account with permissions to create:
  - ECR repositories
  - S3 buckets
  - IAM roles (if needed)

## Usage

### Initialize Terraform

```bash
cd terraform
terraform init
```

### Plan Changes

```bash
terraform plan
```

### Apply Configuration

```bash
terraform apply
```

### Customize Variables

Create a `terraform.tfvars` file:

```hcl
aws_region         = "us-west-2"
environment        = "prod"
bucket_name_prefix = "my-company"
```

Or pass variables via command line:

```bash
terraform apply -var="aws_region=us-west-2" -var="environment=prod"
```

## Outputs

After applying, Terraform will output:
- `ecr_repository_url`: URL of the ECR repository (use this for pushing images)
- `s3_bucket_name`: Name of the S3 bucket for data
- `s3_bucket_arn`: ARN of the S3 bucket

## Example: Push Image to ECR

```bash
# Get ECR login token
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Tag image
docker tag pubmed-api:local <ecr_repository_url>:latest

# Push image
docker push <ecr_repository_url>:latest
```

## Example: Upload Data to S3

```bash
# Upload sample data
aws s3 cp ../data/sample_100_pubmed.jsonl s3://<s3_bucket_name>/pubmed.jsonl
```

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

**Warning**: This will delete the ECR repository and all images, as well as the S3 bucket and all data. Make sure to backup any important data first.

## Next Steps

After provisioning infrastructure:
1. Build and push Docker image to ECR
2. Upload article data to S3
3. Deploy application using ECS Fargate or App Runner (see `docs/aws-deploy.md`)
4. Configure IAM roles for ECS tasks to access S3

## Notes

- The ECR repository is configured with image scanning enabled
- S3 bucket has versioning and encryption enabled
- S3 bucket has public access blocked for security
- ECR lifecycle policy keeps the last 10 images

For production deployments, consider adding:
- ECS cluster and service definitions
- Application Load Balancer
- CloudWatch log groups
- IAM roles and policies
- VPC and networking configuration

