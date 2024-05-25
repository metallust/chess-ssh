setup:
	go install github.com/cosmtrek/air@latest

build:
	go build -o bin/main cmd/main.go

brun:
	go build -o bin/main cmd/main.go
	./bin/main

run:
	go run cmd/main.go
