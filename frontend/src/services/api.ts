import type { Stream, CreateStreamPayload, APIResponse } from '../types/stream';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
const WS_BASE_URL = import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8080';

export async function fetchStreams(): Promise<Stream[]> {
  const res = await fetch(`${API_BASE_URL}/api/streams`);
  if (!res.ok) {
    throw new Error(`Failed to fetch streams: ${res.statusText}`);
  }
  const json: APIResponse<Stream[]> = await res.json();
  return json.data || [];
}

export async function createStream(payload: CreateStreamPayload): Promise<Stream> {
  const res = await fetch(`${API_BASE_URL}/api/streams`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  const json: APIResponse<Stream> = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.error || 'Failed to create stream');
  }
  return json.data!;
}

export async function deleteStream(id: string): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/api/streams/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const json = await res.json().catch(() => ({}));
    throw new Error(json.error || 'Failed to delete stream');
  }
}

export function getStreamWebSocketUrl(streamId: string): string {
  return `${WS_BASE_URL}/ws/stream/${streamId}`;
}
