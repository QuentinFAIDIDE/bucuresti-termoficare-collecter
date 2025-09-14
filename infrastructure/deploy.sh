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
REPO_URI=$(aws ecr describe-repositories --repository-names ${ENVIRONMENT}-termoficare --query 'repositories[0].repositoryUri' --output text 2>/dev/null || echo "")

if [ -z "$REPO_URI" ]; then
    echo "ECR repository not found. Deploy CDK first to create it."
    exit 1
fi

# Login to ECR
aws ecr get-login-password --region $(aws configure get region) | podman login --username AWS --password-stdin $REPO_URI

# Build and push ETL Lambda
podman build -f ../etl_lambda.Dockerfile -t $REPO_URI:etl-$VERSION_TAG ..
podman push $REPO_URI:etl-$VERSION_TAG
podman build -f ../etl_lambda.Dockerfile -t $REPO_URI:etl-latest ..
podman push $REPO_URI:etl-latest

# Build and push API Lambdas
podman build -f ../get_counts_lambda.Dockerfile -t $REPO_URI:api-getcounts-$VERSION_TAG ..
podman push $REPO_URI:api-getcounts-$VERSION_TAG
podman build -f ../get_counts_lambda.Dockerfile -t $REPO_URI:api-getcounts-latest ..
podman push $REPO_URI:api-getcounts-latest

podman build -f ../get_stations_lambda.Dockerfile -t $REPO_URI:api-getstations-$VERSION_TAG ..
podman push $REPO_URI:api-getstations-$VERSION_TAG
podman build -f ../get_stations_lambda.Dockerfile -t $REPO_URI:api-getstations-latest ..
podman push $REPO_URI:api-getstations-latest

podman build -f ../get_station_details_lambda.Dockerfile -t $REPO_URI:api-getstationdetails-$VERSION_TAG ..
podman push $REPO_URI:api-getstationdetails-$VERSION_TAG
podman build -f ../get_station_details_lambda.Dockerfile -t $REPO_URI:api-getstationdetails-latest ..
podman push $REPO_URI:api-getstationdetails-latest

echo "Images pushed to $REPO_URI:"
echo "ETL: etl-$VERSION_TAG and etl-latest"
echo "API GetCounts: api-getcounts-$VERSION_TAG and api-getcounts-latest"
echo "API GetStations: api-getstations-$VERSION_TAG and api-getstations-latest"
echo "API GetStationDetails: api-getstationdetails-$VERSION_TAG and api-getstationdetails-latest"