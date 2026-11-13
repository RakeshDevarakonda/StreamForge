# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and compile
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Stage 2: Minimal runtime image with FFmpeg installed
FROM alpine:3.20

WORKDIR /app

# Install FFmpeg and SSL certificates
RUN apk add --no-cache ffmpeg ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/server .

# Default production environment
ENV PORT=8088
ENV ENV=production
ENV ALLOWED_ORIGINS=*

EXPOSE 8088

CMD ["./server"]
