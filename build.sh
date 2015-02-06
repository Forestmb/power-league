#!/bin/bash
set -e
# To run this before every commit, use: 
#     $ ln -s "$(pwd)/build.sh" .git/hooks/pre-commit
 
dir="$(dirname "$(readlink -f "$0")")"
cd "${dir}"

export PATH="${GOPATH}/bin:${PATH}"

# List known golint warnings that are being suppressed
#
#   1. Consistent with OAuth 1.0 Consumer type
lint_exclusions=(
    "func GetRequestTokenAndUrl should be GetRequestTokenAndURL" 
)

echo "Running go get..."
go get

echo "Running golint..."
ignore=
for exclusion in "${lint_exclusions[@]}"; do
    if [ -z "${ignore}" ]
    then
        ignore="${exclusion}"
    else
        ignore="${exclusion}|${ignore}"
    fi
done
go get github.com/golang/lint/golint
if golint ./... | grep -v -E "${ignore}" | grep .
then
    echo "golint: warnings detected " 1>&2
    echo "False positives should be added to the exclusion list in 'build.sh'" 1>&2
    exit 2
fi

echo "Running go vet..."
go vet ./...

echo "Running goimports..."
go get code.google.com/p/go.tools/cmd/goimports
goimports -w .

echo "Running go fmt..."
go fmt ./...

echo "Running tests..."
# Snippet taken from https://gist.github.com/hailiang/0f22736320abe6be71ce
echo "mode: count" > profile.cov
for dir in $(find . -maxdepth 10 -not -path './.git*' -not -path '*/_*' -type d);
do
if ls "${dir}/"*.go &> /dev/null; then
    go test -v -covermode=count -coverprofile="${dir}/profile.tmp" "${dir}"
    if [ -f "${dir}/profile.tmp" ]
    then
        tail -n +2 "${dir}/profile.tmp" >> profile.cov
        rm "${dir}/profile.tmp"
    fi
fi
done
go tool cover -func profile.cov

echo "Building binary..."
go build .
