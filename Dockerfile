# Multi-stage build: React frontend → Go binary (with frontend embedded
# via //go:embed all:frontend/dist) → minimal runtime image.

# Global build arg — must be declared BEFORE any FROM line so it's visible
# in the FROM that uses it. fly.toml passes this via [build.args].
ARG GO_VERSION=1.25.0

# --- 1. Build the React frontend ---
FROM node:20-bookworm AS frontend
WORKDIR /usr/src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# --- 2. Build the Go binary ---
FROM golang:${GO_VERSION}-bookworm AS backend
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
# Drop in the freshly-built frontend/dist so go:embed picks it up.
COPY --from=frontend /usr/src/frontend/dist ./frontend/dist
RUN go build -v -o /run-app .

# --- 3. Runtime image ---
# The Go binary reads `templates/*.html` via ParseGlob and serves
# `static/` via http.FileServer at /static/* — both are loaded from the
# CWD at runtime, so we set WORKDIR=/app and copy them in.
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend /run-app /app/run-app
COPY templates ./templates
COPY static ./static
CMD ["/app/run-app"]
