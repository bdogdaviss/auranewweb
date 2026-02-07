.PHONY: run build deps clean

deps:
	go mod download

run: deps
	go run main.go

build: deps
	go build -o bin/aura-server main.go

clean:
	rm -rf bin/
