import json
import boto3
import gzip
from datetime import datetime
import os

s3 = boto3.client('s3')
BUCKET_NAME = os.environ['BACKUP_BUCKET']

def lambda_handler(event, context):
    records = []
    
    for record in event['Records']:
        if record['eventName'] in ['INSERT', 'MODIFY', 'REMOVE']:
            # Extract the new image from DynamoDB stream
            dynamodb_record = record['dynamodb']
            
            # Convert DynamoDB format to regular JSON
            if 'NewImage' in dynamodb_record:
                item = {}
                for key, value in dynamodb_record['NewImage'].items():
                    if 'S' in value:
                        item[key] = value['S']
                    elif 'N' in value:
                        item[key] = int(value['N']) if '.' not in value['N'] else float(value['N'])
                    elif 'BOOL' in value:
                        item[key] = value['BOOL']
                
                records.append({
                    'timestamp': record['dynamodb']['ApproximateCreationDateTime'],
                    'item': item
                })
    
    if records:
        # Create S3 key with daily partitioning
        now = datetime.utcnow()
        s3_key = f"{now.strftime('%Y-%m-%d')}/batch_{now.strftime('%Y%m%d_%H%M%S')}.json.gz"
        
        # Compress and upload to S3
        json_data = json.dumps(records)
        compressed_data = gzip.compress(json_data.encode('utf-8'))
        
        s3.put_object(
            Bucket=BUCKET_NAME,
            Key=s3_key,
            Body=compressed_data,
            ContentType='application/gzip'
        )
        
        print(f"Uploaded {len(records)} records to s3://{BUCKET_NAME}/{s3_key}")
    
    print(f"Successfully processed {len(records)} records")
    return