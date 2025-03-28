# Code Refactor Tool

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
- `REPO_URL` (Required) - GitHub repository URL
- `GITHUB_TOKEN` (Required) - GitHub authentication token

Example:
```sh
export REPO_URL="https://github.com/example/repo.git"
export GITHUB_TOKEN="your_github_token"
./code-refactoring-tool
```

### Testing
Run unit tests with:
```sh
make ci
```
