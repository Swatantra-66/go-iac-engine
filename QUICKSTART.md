# Quickstart Guide

Get the Go-IaC-Engine up and running on your local machine quickly.

## 1. Prerequisites

- Go 1.20+ installed.
- An AWS Account with programmatic access (Access Key & Secret Key).
- AWS SDK v2 installed
- Postman or curl for API testing.

## 2. Clone & Setup

`bash
git clone https://github.com/Swatantra-66/go-iac-engine.git
cd go-iac-engine
`

## 3. Configure Secrets

Create a .env file in the root directory and add your credentials:

`text
IAC_API_KEY=your-custom-secret-key
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_REGION=us-east-1
OUTPUT_FORMAT=json
`

## 4. Build & Run

`bash
chmod +x deploy.sh
./deploy.sh
./iac-engine
`

The server will start on http://localhost:8080.

## 5. Deploy Infrastructure

Send a POST request to provision the resources defined in main.yaml.

`bash
curl -X POST http://localhost:8080/deploy \
     -H "X-API-Key: your-custom-secret-key" \
     -H "Content-Type: application/json"
`

## 6. Test Drift Detection (Update)

1. Open main.yaml and change the EC2 InstanceType from t2.micro to t2.small.
2. Run the POST /deploy request again.
3. The engine will detect the drift, safely stop the instance, apply the hardware upgrade using AWS Waiters, and restart it.

## 7. Teardown

To avoid AWS charges, ensure you destroy your resources when finished:

`bash
curl -X DELETE http://localhost:8080/destroy \
     -H "X-API-Key: your-custom-secret-key"
`
