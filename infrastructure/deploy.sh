#!/bin/bash
set -e


# Check if arguments are passed
if [ $# -lt 2 ]; then
    echo "Error: Missing arguments"
    echo "Usage: $0 <version_tag> <environment>"
    echo "Example: $0 v1.0.0 dev"
    exit 1
fi

VERSION_TAG=$1
ENVIRONMENT=$2

# Get ECR repository URI
REPO_URI=$(aws ecr describe-repositories --repository-names ${ENVIRONMENT}-bucuresti-termoficare-lambda --query 'repositories[0].repositoryUri' --output text 2>/dev/null || echo "")

if [ -z "$REPO_URI" ]; then
    echo "ECR repository not found. Deploy CDK first to create it."
    exit 1
fi

# Login to ECR
aws ecr get-login-password --region $(aws configure get region) | podman login --username AWS --password-stdin $REPO_URI

# Build and push Docker image
podman build -f ../etl_lambda.Dockerfile -t $REPO_URI:$VERSION_TAG ..
podman push $REPO_URI:$VERSION_TAG

podman build -f ../etl_lambda.Dockerfile -t $REPO_URI:latest ..
podman push $REPO_URI:latest

echo "Image pushed to $REPO_URI:$VERSION_TAG and $REPO_URI:latest"