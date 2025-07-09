# also run tests under ./infra/stack
test:
	@echo "Running tests..."
	@go test ./...
	@cd infra && go test ./...
	@echo "Tests passed."

lint:
	@echo "Running linter..."
	@golangci-lint -v run
	@cd infra/stack && golangci-lint -v run
	@echo "Linter passed."

mock:
	@echo "Generating mocks..."
	@go generate ./...
	@echo "Mocks generated."

ci: mock test lint

bootstrap:
	@echo "Running CDK bootstrap..."
	@cd infra && cdk bootstrap