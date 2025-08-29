# Run all backend tests with verbose output
test:
	go test -v ./...

# Format Go code using goimports
format:
	goimports -local github.com/aliskhannn/order-service -w .

# Run linters: vet + golangci-lint
lint:
	go vet ./... && golangci-lint run ./...

# Build and start all Docker services
docker-up:
	docker compose up --build

# Stop and remove all Docker services and volumes
docker-down:
	docker compose down -v