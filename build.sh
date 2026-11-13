#!/usr/bin/env bash
set -e

echo "=== Installing static FFmpeg for Render Native Environment ==="

# Create local bin directory inside backend so it sits right next to the compiled server
mkdir -p backend/bin

# Download official static Linux x86_64 FFmpeg build
curl -sL https://github.com/eugeneware/ffmpeg-static/releases/download/b5.0.1/ffmpeg-linux-x64 -o backend/bin/ffmpeg
chmod +x backend/bin/ffmpeg

echo "=== FFmpeg installed successfully ==="

echo "=== Building Go server ==="
cd backend
go build -o server ./cmd/server

echo "=== Build Complete ==="
