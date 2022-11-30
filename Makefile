test:
	@go test ./... -coverpkg=./... -race -coverprofile=coverage.out -covermode=atomic

lint:
	golangci-lint run
