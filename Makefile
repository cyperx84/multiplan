.PHONY: build test lint install clean

BINARY := multiplan

build:
	go build -o $(BINARY) .

test:
	go test -cover -race ./...

lint:
	go vet ./...

install:
	go install .

clean:
	rm -f $(BINARY)
