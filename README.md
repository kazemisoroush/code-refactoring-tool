# Code Refactor Tool

A Go-based API service for automated code refactoring using AI.

## Infrastructure

The AWS infrastructure for this project is maintained in a separate repository:
[code-refactoring-infra](https://github.com/kazemisoroush/code-refactoring-infra)

## How to Run

### Prerequisites
- Install [Docker](https://docs.docker.com/get-docker/)

### Run with Docker
```sh
# Build the Docker image
docker build -t code-refactoring-tool .

# Run the container
docker run --rm -p 8080:8080 code-refactoring-tool
```

### Run Locally
```sh
# Ensure Go 1.24.1 is installed
# Install dependencies
go mod tidy

# Build the application
go build -o code-refactoring-tool ./main.go

# Run the application
./code-refactoring-tool
```

### Environment Variables
This application uses environment variables for configuration:
- `GIT_REPO_URL` (Required) - GitHub repository URL
- `GIT_TOKEN` (Required) - GitHub authentication token

Example:
```sh
export GIT_REPO_URL="https://github.com/example/repo.git"
export GIT_TOKEN="your_github_token"
./code-refactoring-tool
```

### Testing
Run unit tests with:
```sh
make test
```

Run linting with:
```sh
make lint
```
