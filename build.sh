#!/bin/bash
set -e
# To run this before every commit, use: 
#     $ ln -s "$(pwd)/build.sh" .git/hooks/pre-commit
 
dir="$(dirname "$(readlink -f "$0")")"
cd "${dir}"

export PATH="${GOPATH}/bin:${PATH}"

echo "Running go get..."
go get

echo "Running tests..."
go test -v ./...

echo "Running golint..."
go get github.com/golang/lint/golint
golint .

echo "Running go vet..."
go get -d code.google.com/p/go.tools/cmd/vet
go vet .

echo "Running goimports..."
go get code.google.com/p/go.tools/cmd/goimports
goimports -w .

echo "Running go fmt..."
go fmt ./...

echo "Building..."
go build .
