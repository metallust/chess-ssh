setup:
	go install github.com/cosmtrek/air@latest

build:
	go build -o bin/chessssh cmd/main.go

run: build
	./bin/chessssh
