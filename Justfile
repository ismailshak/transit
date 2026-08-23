# just defaults to `sh` so we need to handle Windows
[windows]
set shell := ["powershell.exe", "-NoLogo", "-Command"]

# List the available recipes
default:
    @just --list

# Build the binary into the repo root
build:
    go build -o transit

# Run the test suite
test:
    go test ./...

# Run the test suite with the race detector
race:
    go test -race ./...

# Run the linters that gate the branch
lint:
    golangci-lint run

# Format every Go file
fmt:
    golangci-lint fmt

# Add missing modules and remove unused modules
tidy:
    go mod tidy

# Regenerate a package's golden files, e.g. `just golden ./internal/render -run TestBoard/two_stations`
golden pkg *args:
    go test {{ pkg }} -update -count=1 {{ args }}

# Run the CLI from source, e.g. `just run at ballston`
run *args:
    go run . {{ args }}
