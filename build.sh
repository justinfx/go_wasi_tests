#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $SCRIPT_DIR

echo "Building Extism wasm plugin (slow)"
tinygo build -o ./plugin/extism/plugin_extism.wasm -target wasi ./plugin/extism/plugin_extism.go

echo "Building native WASI wasm plugin"
GOOS=wasip1 GOARCH=wasm go build -o ./plugin/plugin.wasm ./plugin/plugin.go

echo "Building main host app"
go build -o ./main ./main.go
