terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# ECR Repository for Docker images
resource "aws_ecr_repository" "pubmed_api" {
  name                 = "pubmed-api"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = {
    Name        = "pubmed-api"
    Environment = var.environment
    Project     = "pubmed-api"
  }
}

# ECR Lifecycle Policy to manage image retention
resource "aws_ecr_lifecycle_policy" "pubmed_api" {
  repository = aws_ecr_repository.pubmed_api.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus     = "any"
          countType     = "imageCountMoreThan"
          countNumber   = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# S3 Bucket for article data
resource "aws_s3_bucket" "pubmed_data" {
  bucket = "${var.bucket_name_prefix}-pubmed-data-${var.environment}"

  tags = {
    Name        = "pubmed-data"
    Environment = var.environment
    Project     = "pubmed-api"
  }
}

# S3 Bucket Versioning
resource "aws_s3_bucket_versioning" "pubmed_data" {
  bucket = aws_s3_bucket.pubmed_data.id

  versioning_configuration {
    status = "Enabled"
  }
}

# S3 Bucket Server-Side Encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "pubmed_data" {
  bucket = aws_s3_bucket.pubmed_data.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# S3 Bucket Public Access Block
resource "aws_s3_bucket_public_access_block" "pubmed_data" {
  bucket = aws_s3_bucket.pubmed_data.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets  = true
}

# Outputs
output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = aws_ecr_repository.pubmed_api.repository_url
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket for data"
  value       = aws_s3_bucket.pubmed_data.id
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = aws_s3_bucket.pubmed_data.arn
}

