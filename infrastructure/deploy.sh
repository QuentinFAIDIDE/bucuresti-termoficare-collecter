#!/bin/bash
set -e


# Check if an argument is passed
if [ $# -eq 0 ]; then
    echo "Error: No argument provided"
    echo "Usage: $0 <version_tag>"
    exit 1
fi

VERSION_TAG=$1

# Build Go Lambda
cd ../etl_lambda
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
cd ../infrastructure

# Get ECR repository URI
REPO_URI=$(aws ecr describe-repositories --repository-names bucuresti-termoficare-lambda --query 'repositories[0].repositoryUri' --output text 2>/dev/null || echo "")

if [ -z "$REPO_URI" ]; then
    echo "ECR repository not found. Deploy CDK first to create it."
    exit 1
fi

# Login to ECR
aws ecr get-login-password --region $(aws configure get region) | docker login --username AWS --password-stdin $REPO_URI

docker build -t $REPO_URI:VERSION_TAG .
docker push $REPO_URI:VERSION_TAG

# Build and push Docker image
docker build -t $REPO_URI:latest .
docker push $REPO_URI:latest

echo "Image pushed to $REPO_URI:latest"