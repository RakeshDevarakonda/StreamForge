# StreamForge 🎥⚡

> High-performance, ultra low-latency multi-camera RTSP live streaming platform built with Go, FFmpeg, React, TypeScript, and WebSockets.

StreamForge bridges the gap between RTSP video feeds (IP cameras, CCTV, NVRs, drones) and modern web browsers. It transmuxes live RTSP streams on-the-fly into lightweight MPEG-TS chunks and streams them via WebSockets with sub-second latency (~100ms) directly onto HTML5 `<canvas>` elements using JSMpeg WebGL/WASM decoding.

---

## 🌐 Live Deployments & URLs

| Service | Environment | URL |
| :--- | :--- | :--- |
| **Frontend Web App** | Production (Vercel) | [https://streamforge-two.vercel.app](https://streamforge-two.vercel.app/) |
| **Backend API / WebSocket** | Production (Render) | [https://streamforge-ij5e.onrender.com](https://streamforge-ij5e.onrender.com/health) |
| **Local Frontend** | Development | `http://localhost:5173` |
| **Local Backend** | Development | `http://localhost:8088` (WebSocket: `ws://localhost:8088/ws/live?url=<rtsp_url>`) |

---

## 🌟 Key Features

- **Sub-Second Playback Latency:** Stream live RTSP feeds with ~100ms end-to-end delay without buffering pauses.
- **100% Stateless Backend:** No database, no in-memory CRUD storage. FFmpeg runs on-demand per stream and stops when disconnected.
- **Multi-Camera Grid Layout:** Simultaneous multi-stream viewing with responsive layouts (Auto, Single Focus, 2x2, 3x3).
- **Client-Side State Persistence:** Camera streams persist cleanly in browser `localStorage`. No data loss on server restarts.
- **Auto-Idle Resource Reclamation:** Automatically terminates idle FFmpeg processes after 30 seconds with 0 viewers to preserve host CPU & memory.
- **Security & Privacy First:** Built-in SSRF protection (blocking cloud metadata IPs), strict URL scheme validation, and automatic credential masking (`rtsp://***:***@host`).
- **Responsive Modern UI:** Fast, mobile-friendly interface built with React 19, TypeScript, and Tailwind CSS.

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
│  • Stateless On-Demand Hub      │
│  • FFmpeg Transcoder (MPEG-TS)  │
│  • WebSocket Fan-Out Streamer   │
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
- **FFmpeg:** Installed on your system:
  - **Windows:** `winget install Gyan.FFmpeg.Essentials`
  - **macOS:** `brew install ffmpeg`
  - **Linux / Ubuntu:** `sudo apt install ffmpeg`

---

### 1. Start the Backend

```bash
cd backend

# Copy sample configuration
cp .env.example .env

# Run server
go run ./cmd/server
```

The Go server starts listening on **`http://localhost:8088`**.

Health check verification:
```bash
curl http://localhost:8088/health
# Response: {"status":"ok","timestamp":"..."}
```

---

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

## 🧪 Testing with RTSP Streams

### Option A: Local Test Stream with MediaMTX & FFmpeg (Recommended)

1. **Download & Start MediaMTX:**
   Download the standalone [MediaMTX binary](https://github.com/bluenviron/mediamtx/releases) and run it:
   ```bash
   ./mediamtx
   ```
   *MediaMTX will bind RTSP to port `8554`.*

2. **Publish a Synthetic Test Camera Feed:**
   Open a separate terminal and run:
   ```bash
   ffmpeg -re -f lavfi -i "color=c=0x0f172a:s=1280x720:r=25" -vf "drawtext=fontfile='C\:/Windows/Fonts/arial.ttf':text='LIVE CAMERA 01':x=(w-text_w)/2:y=180:fontsize=48:fontcolor=0x38bdf8,drawtext=fontfile='C\:/Windows/Fonts/arial.ttf':text='TIME\: %{pts\:hms}':x=(w-text_w)/2:y=380:fontsize=72:fontcolor=white:box=1:boxcolor=0x000000@0.8:boxborderw=15" -pix_fmt yuv420p -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp -rtsp_transport tcp rtsp://localhost:8554/live/camera1
   ```

3. **View in StreamForge:**
   In the StreamForge UI, paste:
   - **Stream Name:** `Camera 01`
   - **RTSP URL:** `rtsp://localhost:8554/live/camera1`

---

### Option B: Public RTSP Stream Feeds

You can also test directly with public RTSP streams:
- `rtsp://stream.strba.sk:1935/strba/live.stream`

---

## 📡 Backend API Reference

The backend is intentionally 100% stateless with zero database requirements:

| Endpoint | Protocol | Description | Example |
| :--- | :--- | :--- | :--- |
| `/health` | HTTP GET | Server health & readiness probe | `http://localhost:8088/health` |
| `/ws/live` | WebSocket | On-demand live binary MPEG-TS video stream | `ws://localhost:8088/ws/live?url=rtsp://localhost:8554/live/camera1` |

---

## ⚙️ Environment Variables

### Backend (`backend/.env`)

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8088` | Port the backend HTTP/WebSocket server listens on |
| `ENV` | `development` | Environment mode (`development` or `production`) |
| `ALLOWED_ORIGINS` | `*` | Comma-separated CORS allowed origins |
| `FFMPEG_PATH` | *(Auto-detected)* | Optional explicit path to FFmpeg binary |

### Frontend (`frontend/.env`)

| Variable | Default | Description |
| :--- | :--- | :--- |
| `VITE_WS_URL` | `ws://localhost:8088` | WebSocket backend URL used by the player |

---

## 📄 License
MIT License.
