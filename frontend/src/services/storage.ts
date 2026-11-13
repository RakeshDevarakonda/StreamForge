const LOCAL_STORAGE_KEY = 'streamforge_saved_cameras';

export interface SavedCamera {
  id: string;
  name: string;
  rtsp_url: string;
}

// Normalize RTSP URL for robust comparison
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

// Save a newly added camera to localStorage (prevents duplicates)
export function saveCameraToStorage(camera: SavedCamera): void {
  try {
    const existing = getSavedCameras();
    const targetNorm = normalizeRtspUrl(camera.rtsp_url);
    const filtered = existing.filter((c) => normalizeRtspUrl(c.rtsp_url) !== targetNorm);
    const updated: SavedCamera[] = [...filtered, camera];
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(updated));
  } catch (err) {
    console.error('Failed to save camera to localStorage:', err);
  }
}

// Remove a camera from localStorage by rtsp_url
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
