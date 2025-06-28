# also run tests under ./infra/stack
test:
	@echo "Running tests..."
	@go test ./...
	@cd infra/stack && go test ./...
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
