all:
	go run ./cmd/static-generator/main.go
	go build -o ./docuowl ./cmd/docuowl/main.go
