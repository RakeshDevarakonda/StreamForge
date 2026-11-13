# StreamForge Frontend

A modern, high-performance RTSP live streaming interface built with React 19, TypeScript, Vite, and Tailwind CSS.

## Features
- **Low-Latency Live Playback:** Streams raw MPEG-TS video over WebSockets with JSMpeg decoding onto HTML5 `<canvas>`.
- **Responsive Multi-Stream Grid:** Flexible layouts (Auto, Single Column, 2x2, 3x3) for viewing multiple camera streams simultaneously.
- **Local State Rehydration:** Saved cameras persist across browser sessions and automatic reconnection occurs seamlessly.
- **Credential Protection:** Automatic masking of sensitive credentials in camera URLs.

## Development
```bash
npm install
npm run dev
```

## Production Build
```bash
npm run build
```
