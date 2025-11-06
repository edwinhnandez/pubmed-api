# AWS Deployment Guide

This document provides guidance for deploying the PubMed API to AWS using **ECS Fargate** or **AWS App Runner**.

## Overview

The application is designed to run locally without AWS dependencies but can be easily deployed to AWS with minimal configuration.

## Prerequisites

- AWS CLI configured with appropriate credentials
- Docker (for building images)
- Terraform (optional, for infrastructure as code)

## Architecture Options

### Option 1: AWS App Runner (Recommended for Simplicity)

**Pros:**
- Fully managed service
- Automatic scaling
- Built-in load balancing
- Simple deployment process

**Cons:**
- Less control over infrastructure
- Limited customization options

### Option 2: ECS Fargate (Recommended for Production)

**Pros:**
- More control and flexibility
- Better for production workloads
- Integration with other AWS services
- Blue/green deployments

**Cons:**
- More complex setup
- Requires VPC, ALB configuration

---

## Deployment: AWS App Runner

### Step 1: Build and Push Docker Image to ECR

```bash
# Create ECR repository
aws ecr create-repository --repository-name pubmed-api

# Get login token
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build image
docker build -t pubmed-api:latest .

# Tag image
docker tag pubmed-api:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/pubmed-api:latest

# Push image
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/pubmed-api:latest
```

### Step 2: Create App Runner Service

**Via AWS Console:**
1. Go to AWS App Runner console
2. Create a new service
3. Choose "Container image"
4. Select your ECR image
5. Configure:
   - **Service name:** `pubmed-api`
   - **Port:** `8080`
   - **Health check path:** `/healthz`
   - **Auto deploy:** Enabled
6. Set environment variables:
   ```
   PORT=8080
   DATA_S3_URL=s3://your-bucket/pubmed.jsonl
   LOG_LEVEL=info
   DB_PATH=/tmp/pubmed.db
   ```
7. Configure IAM role with S3 read permissions
8. Create service

**Via AWS CLI:**
```bash
aws apprunner create-service \
  --service-name pubmed-api \
  --source-configuration '{
    "ImageRepository": {
      "ImageIdentifier": "<account-id>.dkr.ecr.us-east-1.amazonaws.com/pubmed-api:latest",
      "ImageConfiguration": {
        "Port": "8080",
        "RuntimeEnvironmentVariables": {
          "PORT": "8080",
          "DATA_S3_URL": "s3://your-bucket/pubmed.jsonl",
          "LOG_LEVEL": "info",
          "DB_PATH": "/tmp/pubmed.db"
        }
      },
      "ImageRepositoryType": "ECR"
    },
    "AutoDeploymentsEnabled": true
  }' \
  --health-check-configuration '{
    "Protocol": "HTTP",
    "Path": "/healthz",
    "Interval": 10,
    "Timeout": 5,
    "HealthyThreshold": 1,
    "UnhealthyThreshold": 5
  }'
```

---

## Deployment: ECS Fargate

### Step 1: Build and Push Docker Image

Same as App Runner Step 1.

### Step 2: Create S3 Bucket for Data

```bash
aws s3 mb s3://pubmed-api-data
aws s3 cp data/sample_100_pubmed.jsonl s3://pubmed-api-data/pubmed.jsonl
```

### Step 3: Create ECS Task Definition

**task-definition.json:**
```json
{
  "family": "pubmed-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "pubmed-api",
      "image": "<account-id>.dkr.ecr.us-east-1.amazonaws.com/pubmed-api:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "PORT",
          "value": "8080"
        },
        {
          "name": "DATA_S3_URL",
          "value": "s3://pubmed-api-data/pubmed.jsonl"
        },
        {
          "name": "LOG_LEVEL",
          "value": "info"
        },
        {
          "name": "DB_PATH",
          "value": "/tmp/pubmed.db"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/pubmed-api",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "wget --quiet --tries=1 --spider http://localhost:8080/healthz || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

Register task definition:
```bash
aws ecs register-task-definition --cli-input-json file://task-definition.json
```

### Step 4: Create CloudWatch Log Group

```bash
aws logs create-log-group --log-group-name /ecs/pubmed-api
```

### Step 5: Create ECS Cluster

```bash
aws ecs create-cluster --cluster-name pubmed-api-cluster
```

### Step 6: Create IAM Role for ECS Task

The task role needs S3 read permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject"
      ],
      "Resource": "arn:aws:s3:::pubmed-api-data/*"
    }
  ]
}
```

### Step 7: Create ECS Service

**Via AWS Console:**
1. Go to ECS Console → Clusters → pubmed-api-cluster
2. Create Service
3. Configure:
   - **Launch type:** Fargate
   - **Task definition:** pubmed-api
   - **Service name:** pubmed-api-service
   - **Number of tasks:** 2 (for high availability)
   - **VPC:** Select your VPC
   - **Subnets:** Select public subnets
   - **Security groups:** Allow inbound on port 8080
   - **Load balancer:** Create Application Load Balancer
   - **Health check:** `/healthz`

**Via AWS CLI:**
```bash
aws ecs create-service \
  --cluster pubmed-api-cluster \
  --service-name pubmed-api-service \
  --task-definition pubmed-api \
  --desired-count 2 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}" \
  --load-balancers "targetGroupArn=arn:aws:elasticloadbalancing:region:account:targetgroup/xxx,containerName=pubmed-api,containerPort=8080"
```

---

## Configuration

### Environment Variables

| Variable | Description | Production Value |
|----------|-------------|------------------|
| `PORT` | HTTP server port | `8080` |
| `DATA_S3_URL` | S3 URL to dataset | `s3://bucket/pubmed.jsonl` |
| `LOG_LEVEL` | Logging level | `info` |
| `DB_PATH` | SQLite database path | `/tmp/pubmed.db` (or EFS mount) |

### Health Checks

- **Path:** `/healthz`
- **Interval:** 30 seconds
- **Timeout:** 5 seconds
- **Healthy threshold:** 1
- **Unhealthy threshold:** 3

### Scaling

**App Runner:**
- Auto-scales based on traffic
- Configure min/max instances in console

**ECS Fargate:**
- Use Auto Scaling based on CPU/Memory utilization
- Target: 70% CPU utilization
- Min: 2 tasks, Max: 10 tasks

---

## Monitoring & Observability

### CloudWatch Logs

Logs are automatically sent to CloudWatch:
- **App Runner:** Service logs in App Runner console
- **ECS:** Log group `/ecs/pubmed-api`

### Metrics

Key metrics to monitor:
- Request count
- Error rate (4xx, 5xx)
- Response time (p50, p95, p99)
- CPU utilization
- Memory utilization

### Alerts

Set up CloudWatch alarms for:
- High error rate (> 5%)
- High response time (p95 > 1s)
- High CPU utilization (> 80%)
- Unhealthy target health checks

---

## Blue/Green Deployments (ECS Fargate)

1. Create new task definition with updated image
2. Use ECS blue/green deployment feature
3. Deploy to new task set
4. Run health checks
5. Switch traffic to new task set
6. Monitor for issues
7. Rollback if needed

---

## Security Best Practices

1. **IAM Roles:** Use task roles (not instance profiles) for S3 access
2. **Secrets:** Use AWS Secrets Manager for sensitive data (if needed)
3. **Network:** Use private subnets with NAT Gateway for ECS tasks
4. **WAF:** Add AWS WAF for DDoS protection
5. **SSL/TLS:** Use ALB with ACM certificate
6. **VPC:** Isolate services in private subnets

---

## Cost Optimization

- **App Runner:** Pay per request and compute time
- **ECS Fargate:** Right-size task CPU/memory
- **S3:** Use S3 Intelligent-Tiering for data storage
- **CloudWatch:** Set log retention policies

---

## Troubleshooting

### Service won't start
- Check CloudWatch logs
- Verify IAM permissions
- Check health check configuration
- Verify S3 object exists and is accessible

### High latency
- Check database performance (consider using EFS for persistent storage)
- Review application logs
- Monitor CloudWatch metrics
- Consider caching layer

### Out of memory
- Increase task memory allocation
- Review database query performance
- Consider using RDS PostgreSQL instead of SQLite

---

## Next Steps

1. **Database Migration:** Move from SQLite to RDS PostgreSQL for production
2. **Caching:** Add ElastiCache (Redis) for query caching
3. **CDN:** Use CloudFront for static assets
4. **API Gateway:** Add API Gateway for rate limiting and API keys
5. **CI/CD:** Set up CodePipeline for automated deployments

---

## Terraform Stub (Optional)

A minimal Terraform configuration is available in `terraform/` directory for:
- ECR repository
- S3 bucket for data
- Basic ECS cluster and service

See `terraform/README.md` for details.

