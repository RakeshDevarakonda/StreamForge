#!/usr/bin/env bash
set -e

echo "=== Installing static FFmpeg for Render Native Environment ==="

# Create local bin directory inside project
mkdir -p "$HOME/bin"

# Download official static Linux x86_64 FFmpeg build (doesn't need root or apt-get)
curl -sL https://github.com/eugeneware/ffmpeg-static/releases/download/b5.0.1/ffmpeg-linux-x64 -o "$HOME/bin/ffmpeg"
chmod +x "$HOME/bin/ffmpeg"

# Add to PATH
export PATH="$HOME/bin:$PATH"

echo "=== FFmpeg installed successfully at $($HOME/bin/ffmpeg -version | head -n 1) ==="

echo "=== Building Go server ==="
cd backend
go build -o server ./cmd/server

echo "=== Build Complete ==="
