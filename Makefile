# Run all backend tests with verbose output
# The Path to Go Modules
GOMOD := $(shell go list -m)

# Run only unit tests (from folders with mocks)
test-unit:
	go test -v -tags=unit ./...

# Run only integration tests (connect to TEST_DATABASE_URL)
test-integration:
	go test -v -tags=integration ./...

# Running tests with coverage report
test-coverage:
	go test -v -coverprofile=coverage.out -tags=unit ./...
	go test -v -coverprofile=coverage.out -tags=integration ./...
	go tool cover -html=coverage.out
	go tool cover -html=coverage_integration.out

# Format Go code using goimports
format:
	goimports -local github.com/aliskhannn/calendar-service -w .

# Run linters: vet + golangci-lint
lint:
	go vet ./... && golangci-lint run ./...

# Build and start all Docker services
docker-up:
	docker compose up --build

# Stop and remove all Docker services and volumes
docker-down:
	docker compose down -v