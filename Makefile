include .env
export $(shell sed 's/=.*//' .env)

run:
	go run ./...

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout 3m --config=.golangci.yml
