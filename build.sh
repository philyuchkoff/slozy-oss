#!/bin/bash

# Build SLOzy для CentOS 9
echo "Building SLOzy for CentOS 9..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o slozy-linux-amd64 main.go

# Уменьшаем размер бинарника
upx --best slozy-linux-amd64

echo "Binary built: slozy-linux-amd64"
ls -lh slozy-linux-amd64
