.PHONY: all build run test clean fmt

MISE := export PATH="$$HOME/.local/bin:$$PATH"; mise exec --

all: build

build:
	$(MISE) go build -o bin/tboard ./cmd/tboard

run:
	$(MISE) go run ./cmd/tboard

test:
	$(MISE) go test -v ./...

fmt:
	$(MISE) go fmt ./...

clean:
	rm -rf bin/
