# Project Structure

This repository follows standard Go project layout conventions to separate the API routing, business logic, and configuration.

```
go-iac-engine/
├── cmd/
│   └── iac/
│       └── main.go          # Application entry point: Gin router and API endpoints
├── internal/
│   ├── parser/
│   │   └── yaml.go          # Reads and parses the declarative main.yaml file
│   ├── provider/
│   │   └── aws.go           # AWS SDK v2 integration (EC2, Waiters, Drift Detection)
│   └── state/
│       └── state.go         # Remote state management (S3 sync and local fallback)
├── main.yaml                # The "Desired State" declarative configuration file
├── deploy.sh                # Automation script for dependency installation and building
├── .env                     # (Ignored) Environment variables: AWS credentials & API keys
├── .gitignore               # Security rules to prevent secret leaks
├── PROJECT_STRUCTURE.md     # Project Architecture
├── QUICKSTART.md            # Quick guide for running it your local machine
├── go.mod & go.sum          # Go module dependencies
├── README.md                # Comprehensive project documentation
└── LICENSE                  # MIT License
```

### Directory Deep Dive

- **cmd/**: Contains the main application. The Gin REST API is initialized in `main.go`, alongside the secure X-API-Key middleware.
- **internal/**: Private application and library code.
  - **parser/yaml.go**: Handles decoding the desired infrastructure state from the user.
  - **provider/aws.go**: Contains the complex lifecycle logic, including the `InstanceStoppedWaiter` for handling EC2 hardware upgrades without race conditions.
  - **state/state.go**: Acts as the "brain" of the engine, ensuring idempotency by comparing the desired YAML state against the actual AWS remote state.
