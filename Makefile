# also run tests under ./infra/stack
test:
	@echo "Running tests..."
	@go test -v ./...
	@cd infra/stack && go test -v ./...
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
