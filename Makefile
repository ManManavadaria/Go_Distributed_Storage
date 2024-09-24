build:
	@go build -o main.go

test:
	@go test ./... -v

run:
	@go run .