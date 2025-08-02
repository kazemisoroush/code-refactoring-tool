# Code Refactor Tool

A Go-based API service for automated code refactoring using AI.

## Infrastructure

The AWS infrastructure for this project is maintained in a separate repository:
[code-refactoring-infra](https://github.com/kazemisoroush/code-refactoring-infra)

# Code Refactor Tool

A Go-based API service for automated code refactoring using AI.

## Infrastructure

The AWS infrastructure for this project is maintained in a separate repository:
[code-refactoring-infra](https://github.com/kazemisoroush/code-refactoring-infra)

## Quick Start

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

### Local Development
```sh
# Start the complete local development environment
make serve

# This will start:
# - PostgreSQL database
# - Ollama (local LLM server)
# - ChromaDB (vector database)  
# - Code Refactoring Tool API

# API will be available at: http://localhost:8080
# Swagger docs at: http://localhost:8080/swagger/index.html
```

### Other Commands
```sh
# Run in background
make serve-detached

# Stop all services
make stop

# View API logs
make logs

# Run tests and linting
make ci

# Build production Docker image
make docker-build
```

### Configuration

Local development uses the `.env` file for configuration. Copy `.env.example` to `.env` and modify as needed:

```sh
cp .env.example .env
```

Key configuration options:
- `LOCAL_AI_ENABLED=true` - Use local AI (Ollama + ChromaDB)
- `LOCAL_AI_ENABLED=false` - Use AWS Bedrock (requires AWS credentials)

### Production Deployment
```sh
# Build the Docker image
docker build -t code-refactoring-tool .

# Run the container with your environment variables
docker run --env-file .env -p 8080:8080 code-refactoring-tool
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
