# IaC-Engine-Tool

A secure Infrastructure as Code (IaC) engine built with Go and the Gin framework. This tool manages the lifecycle of AWS resources using a declarative approach and remote state management.

## Features

- **Multi-Resource Support**: Provisions EC2 instances and S3 buckets via AWS SDK v2.
- **Remote State Management**: Uses Amazon S3 as a backend to track infrastructure state and ensure consistency.
- **Drift Detection & Lifecycle**: Implements intelligent updates for EC2 instances using AWS Waiters to manage state transitions.
- **API Security**: Secured with custom Gin middleware and X-API-Key authentication.
- **Environment Isolation**: Manages sensitive credentials using .env and godotenv.

## Tech Stack

- Language: Go (Golang)
- Framework: Gin Gonic
- Cloud: AWS (EC2, S3)
- Configuration: YAML
- Tools: Postman (Testing), AWS SDK v2

## API Usage

Provision or Update:
curl -X POST http://localhost:8080/deploy \
 -H "X-API-Key: your-secret-password" \
 -H "Content-Type: application/json"

Tear Down:
curl -X DELETE http://localhost:8080/destroy \
 -H "X-API-Key: your-secret-password" \
 -H "Content-Type: application/json"
