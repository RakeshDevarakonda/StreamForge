import type { CreateStreamPayload } from '../types/stream';

const LOCAL_STORAGE_KEY = 'streamforge_saved_cameras';

export interface SavedCamera {
  name: string;
  rtsp_url: string;
  saved_at: string;
}

// Normalize RTSP URL for robust comparison (handles masked vs unmasked credentials)
export function normalizeRtspUrl(raw: string): string {
  try {
    const u = new URL(raw);
    return `${u.protocol}//${u.host}${u.pathname}${u.search}`;
  } catch {
    return raw.replace(/\/\/[^@]+@/, '//').trim();
  }
}

// Get all saved cameras from browser localStorage
export function getSavedCameras(): SavedCamera[] {
  try {
    const raw = localStorage.getItem(LOCAL_STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

// Save a newly added camera to localStorage (prevents duplicates by normalized rtsp_url)
export function saveCameraToStorage(camera: CreateStreamPayload): void {
  try {
    const existing = getSavedCameras();
    const targetNorm = normalizeRtspUrl(camera.rtsp_url);
    const filtered = existing.filter((c) => normalizeRtspUrl(c.rtsp_url) !== targetNorm);
    const updated: SavedCamera[] = [
      ...filtered,
      {
        name: camera.name,
        rtsp_url: camera.rtsp_url,
        saved_at: new Date().toISOString(),
      },
    ];
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(updated));
  } catch (err) {
    console.error('Failed to save camera to localStorage:', err);
  }
}

// Remove a camera from localStorage by normalized rtsp_url
export function removeCameraFromStorage(rtspUrl: string): void {
  try {
    const existing = getSavedCameras();
    const targetNorm = normalizeRtspUrl(rtspUrl);
    const updated = existing.filter((c) => normalizeRtspUrl(c.rtsp_url) !== targetNorm);
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(updated));
  } catch (err) {
    console.error('Failed to remove camera from localStorage:', err);
  }
}
