# also run tests under ./infra/stack
test:
	@echo "Running tests..."
	@cd infra/lambda && pytest .
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

deploy:
	@echo "Deploying infra..."
	@cd infra && cdk bootstrap
	@cd infra && cdk deploy --require-approval never
	@echo "Infra deploy done."

destroy:
	@echo "Destroying infra..."
	@cd infra && cdk destroy --all --force
	@echo "Infra destroy done."
