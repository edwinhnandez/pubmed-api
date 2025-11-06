# Outputs are defined in main.tf
# This file is included for completeness and can be used for additional outputs

output "repository_uri" {
  description = "Full URI of the ECR repository"
  value       = aws_ecr_repository.pubmed_api.repository_url
}

output "data_bucket_name" {
  description = "S3 bucket name for article data"
  value       = aws_s3_bucket.pubmed_data.id
}

