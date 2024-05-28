setup:
	go install github.com/cosmtrek/air@latest

build:
	go build -o bin/main cmd/main.go

run: build
	./bin/main	
