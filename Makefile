.PHONY: run build deps clean frontend frontend-deps

deps:
	go mod download

frontend-deps:
	cd frontend && npm install

frontend: frontend-deps
	cd frontend && npm run build

# `make run` rebuilds the React bundle so go:embed picks up the latest changes,
# then starts the Go server. If you're iterating on the frontend specifically
# you'll likely want `cd frontend && npm run dev` in a separate terminal — that
# gets you HMR and proxies API calls to this server.
run: deps frontend
	go run main.go

build: deps frontend
	go build -o bin/aura-server main.go

clean:
	rm -rf bin/ frontend/dist/
