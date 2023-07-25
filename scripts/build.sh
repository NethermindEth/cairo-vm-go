#!/bin/sh

# go to project root
cd "$(dirname "$0")/.."

# create bin dir
mkdir -p bin

# clean previous build
rm -f bin/cairo-vm

# build the new bin
go build -o bin/cairo-vm cmd/cli/main.go

# check for success
if [ $? -eq 0 ]; then
    echo "Build completed succesfully!"
else
    echo "Build failed."
    exit 1
fi
