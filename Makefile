build:
	go build -ldflags="-extldflags=-static -s -w" -o hub .

test:
	@go test ./... -coverpkg=./... -race -coverprofile=coverage.out -covermode=atomic

lint:
	golangci-lint run
