const WS_BASE_URL = import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8088';

// Build on-demand live WebSocket streaming URL with the RTSP feed as parameter
export function getStreamWebSocketUrl(rtspUrl: string): string {
  return `${WS_BASE_URL}/ws/live?url=${encodeURIComponent(rtspUrl)}`;
}
