#!/bin/bash

echo "Installing Go dependencies..."
go mod tidy

if [ ! -f .env ]; then
    echo "Creating .env template..."
    echo "IAC_API_KEY=your-secret-123" > .env
    echo "AWS_ACCESS_KEY_ID=" >> .env
    echo "AWS_SECRET_ACCESS_KEY=" >> .env
    echo "Please fill in your AWS credentials in the .env file."
fi

echo "Building the IaC Engine..."
go build -o iac-engine cmd/iac/main.go

echo "Setup complete. Run './iac-engine' to start the server."