# StreamForge 🎥⚡

> High-performance, ultra low-latency multi-camera RTSP live streaming platform built with Go, FFmpeg, React, TypeScript, and WebSockets.

StreamForge bridges the gap between RTSP video feeds (IP cameras, CCTV, NVRs, drones) and modern web browsers. It transmuxes live RTSP streams on-the-fly into lightweight MPEG-TS chunks and streams them via WebSockets with sub-second latency (~100ms) directly onto HTML5 `<canvas>` elements using JSMpeg WebGL/WASM decoding.

---

## 🌟 Key Features

- **Sub-Second Playback Latency:** Stream live RTSP feeds with ~100ms end-to-end delay without buffering pauses.
- **Multi-Camera Grid Layout:** Simultaneous multi-stream viewing with responsive layouts (Auto, Single Focus, 2x2, 3x3).
- **Resource-Efficient Fan-Out Architecture:** 1 FFmpeg process per camera shared across any number of connected viewers, minimizing CPU & network usage.
- **Auto-Idle Resource Reclamation:** Automatically terminates idle FFmpeg processes after 30 seconds with 0 viewers to preserve host CPU & memory.
- **Resilient Connection Recovery:** Handles sudden camera disconnects with exponential backoff auto-reconnection.
- **Zero-Config Persistent Feeds:** Integrated browser `localStorage` camera rehydration ensures streams seamlessly resume even after server reboots or browser refreshes.
- **Security & Privacy First:** Built-in SSRF protection (blocking cloud metadata IPs), strict URL scheme validation, and automatic credential masking (`rtsp://***:***@host`).
- **Modern Tech Stack:** Go 1.22+ standard HTTP routing, Gorilla WebSocket, React 19, TypeScript, and Tailwind CSS.

---

## 🏗️ Architecture

```
┌─────────────────────────────────┐
│ IP Camera / RTSP Source / Video │
└────────────────+────────────────┘
                 │ (RTSP / TCP)
                 ▼
┌─────────────────────────────────┐
│       StreamForge Backend       │
│           (Go 1.22+)            │
│                                 │
│  • Stream Store (In-Memory)     │
│  • FFmpeg Worker (MPEG-TS)      │
│  • WebSocket Fan-Out Hub        │
└────────────────+────────────────┘
                 │ (Binary MPEG-TS over WS)
                 ▼
┌─────────────────────────────────┐
│       StreamForge Frontend      │
│     (React + TypeScript)        │
│                                 │
│  • JSMpeg WebGL/WASM Decoder    │
│  • HTML5 Canvas Video Player    │
│  • Responsive Grid Dashboard    │
└─────────────────────────────────┘
```

---

## 🚀 Quick Start (Local Setup)

### Prerequisites

- **Go:** `1.22+` installed
- **Node.js:** `v18+` and `npm`
- **FFmpeg:** Installed on your system (`winget install Gyan.FFmpeg.Essentials` on Windows, or `brew install ffmpeg` / `apt install ffmpeg`)

### 1. Start the Backend

```bash
cd backend

# Copy sample configuration
cp .env.example .env

# Run server
go run ./cmd/server
```
The Go server starts listening on `http://localhost:8088`.

### 2. Start the Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start Vite development server
npm run dev
```
Open **`http://localhost:5173`** in your browser.

---

## 🧪 Testing with Local RTSP Streams

You can stream any public RTSP URL or set up your own private local test streams using **MediaMTX** and **FFmpeg**:

### 1. Run MediaMTX
Download the standalone [MediaMTX binary](https://github.com/bluenviron/mediamtx/releases) and execute it:
```bash
./mediamtx
```

### 2. Broadcast Live Video Feeds
Open another terminal and publish test video feeds:

```bash
# Camera 1 (Live running timestamp & frame counter)
ffmpeg -re -f lavfi -i "color=c=0x0f172a:s=1280x720:r=25" -vf "drawtext=fontfile='C\:/Windows/Fonts/arial.ttf':text='LIVE CAMERA 01':x=(w-text_w)/2:y=180:fontsize=48:fontcolor=0x38bdf8,drawtext=fontfile='C\:/Windows/Fonts/arial.ttf':text='TIME\: %{pts\:hms}':x=(w-text_w)/2:y=380:fontsize=72:fontcolor=white:box=1:boxcolor=0x000000@0.8:boxborderw=15" -pix_fmt yuv420p -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp -rtsp_transport tcp rtsp://localhost:8554/live/camera1

# Camera 2 (SMPTE Color Bars)
ffmpeg -re -f lavfi -i smptebars=size=1280x720:rate=25 -pix_fmt yuv420p -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp -rtsp_transport tcp rtsp://localhost:8554/live/camera2
```

In the StreamForge UI, paste:
- `rtsp://localhost:8554/live/camera1`
- `rtsp://localhost:8554/live/camera2`

---

## 📡 REST API Reference

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Server health check |
| `GET` | `/api/streams` | List all active streams |
| `GET` | `/api/streams/{id}` | Get specific stream details |
| `POST` | `/api/streams` | Register a new RTSP stream |
| `DELETE` | `/api/streams/{id}` | Terminate and remove a stream |
| `WS` | `/ws/stream/{id}` | Live binary video WebSocket stream |

---

## 📦 Deployment (Docker)

```bash
# Build & run backend container
docker build -t streamforge-backend ./backend
docker run -p 8088:8088 streamforge-backend
```

---

## 📄 License
MIT License.
