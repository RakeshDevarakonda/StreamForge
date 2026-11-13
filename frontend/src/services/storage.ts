const LOCAL_STORAGE_KEY = 'streamforge_saved_cameras';

export interface SavedCamera {
  id: string;
  name: string;
  rtsp_url: string;
}

// Normalize RTSP URL for strict duplicate prevention
// Strips credentials, trailing slashes, and lowercases for uniform comparison
export function normalizeRtspUrl(raw: string): string {
  if (!raw) return '';
  let cleaned = raw.trim();
  // Strip trailing slashes
  cleaned = cleaned.replace(/\/+$/, '');
  // Strip user:password@ credentials
  cleaned = cleaned.replace(/\/\/[^@]+@/, '//');
  // Lowercase scheme & host
  try {
    const u = new URL(cleaned);
    return `${u.protocol.toLowerCase()}//${u.host.toLowerCase()}${u.pathname.replace(/\/+$/, '')}${u.search}`;
  } catch {
    return cleaned.toLowerCase();
  }
}

// Get all saved cameras from browser localStorage (de-duplicates on read)
export function getSavedCameras(): SavedCamera[] {
  try {
    const raw = localStorage.getItem(LOCAL_STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];

    // Ensure no duplicates exist in storage
    const seen = new Set<string>();
    const unique: SavedCamera[] = [];
    for (const item of parsed) {
      if (!item || !item.rtsp_url) continue;
      const norm = normalizeRtspUrl(item.rtsp_url);
      if (!seen.has(norm)) {
        seen.add(norm);
        unique.push(item);
      }
    }
    return unique;
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
