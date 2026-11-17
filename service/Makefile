.PHONY: build run test lint clean

build:
	go build -o starapp .

run: build
	./starapp -configdir ""

test:
	go test ./...

lint:
	go fmt ./...
	go vet ./...
	golangci-lint run ./...

clean:
	rm -f starapp
	go clean -cache
