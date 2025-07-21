# also run tests under ./infra/stack
test:
	@echo "Running tests..."
	@cd infra/rds_schema_lambda && pytest .
	@go test -v ./...
	@cd infra && go test -v ./...
	@echo "Tests passed."

lint:
	@echo "Running linter..."
	@golangci-lint -v run
	@cd infra/stack && golangci-lint -v run
	@echo "Running Python linter..."
	@cd infra/rds_schema_lambda && pylint --rcfile=../../.pylintrc *.py
	@echo "Linter passed."

mock:
	@echo "Generating mocks..."
	@go generate ./...
	@echo "Mocks generated."

swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go -o docs/
	@echo "Swagger documentation generated."

build-api:
	@echo "Building API server..."
	@go build -o bin/api-server ./cmd/api/
	@echo "API server built."

run-api: swagger build-api
	@echo "Running API server..."
	@./bin/api-server

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf docs/
	@echo "Clean completed."

deploy:
	@echo "Deploying infra..."
	@cd infra && cdk bootstrap
	@cd infra && cdk deploy --require-approval never
	@echo "Infra deploy done."

destroy:
	@echo "Destroying infra..."
	@cd infra && cdk destroy --all --force
	@aws secretsmanager delete-secret \
		--secret-id code-refactor-db-secret \
		--force-delete-without-recovery
	@echo "Infra destroy done."

ci: mock test lint clean swagger
