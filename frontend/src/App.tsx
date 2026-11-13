import React, { useState, useEffect } from 'react';
import { Radio, Plus } from 'lucide-react';
import type { CameraStream } from './types/stream';
import { getSavedCameras, saveCameraToStorage, removeCameraFromStorage } from './services/storage';
import { StreamGrid } from './components/StreamGrid';
import type { GridLayout } from './components/StreamGrid';

export const App: React.FC = () => {
  const [streams, setStreams] = useState<CameraStream[]>([]);
  const [layout, setLayout] = useState<GridLayout>('auto');
  const [name, setName] = useState('');
  const [url, setUrl] = useState('');
  const [error, setError] = useState<string | null>(null);

  // Load saved cameras once on startup
  useEffect(() => {
    const saved = getSavedCameras();
    setStreams(saved);
  }, []);

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const trimmed = url.trim();
    if (!trimmed.startsWith('rtsp://') && !trimmed.startsWith('rtsps://')) {
      setError('URL must start with rtsp:// or rtsps://');
      return;
    }

    const newCam: CameraStream = {
      id: `cam-${Date.now()}`,
      name: name.trim() || `Camera ${streams.length + 1}`,
      rtsp_url: trimmed,
    };

    saveCameraToStorage(newCam);
    setStreams((prev) => [...prev, newCam]);
    setName('');
    setUrl('');
  };

  const handleDelete = (id: string) => {
    const target = streams.find((s) => s.id === id);
    if (target) {
      removeCameraFromStorage(target.rtsp_url);
    }
    setStreams((prev) => prev.filter((s) => s.id !== id));
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans">
      {/* Header */}
      <header className="border-b border-slate-800 px-6 py-3.5 flex items-center justify-between">
        <div className="flex items-center space-x-2.5">
          <Radio className="w-5 h-5 text-cyan-400" />
          <h1 className="font-bold text-base tracking-tight text-white">StreamForge</h1>
          <span className="text-[10px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
            RTSP Viewer
          </span>
        </div>
      </header>

      {/* Main */}
      <main className="max-w-6xl w-full mx-auto p-4 sm:p-6 space-y-6 flex-1">
        {/* Stream Input Form */}
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 shadow-md">
          <form onSubmit={handleAdd} className="flex flex-col sm:flex-row gap-2">
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Camera Name (Optional)"
              className="sm:w-1/3 px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm text-slate-200 focus:outline-none focus:border-cyan-500"
            />
            <input
              type="text"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="rtsp://your-stream-url..."
              className="flex-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm font-mono text-slate-200 focus:outline-none focus:border-cyan-500"
            />
            <button
              type="submit"
              className="px-4 py-2 bg-cyan-500 hover:bg-cyan-400 active:bg-cyan-600 text-slate-950 font-semibold text-sm rounded-lg flex items-center justify-center space-x-1 shrink-0 cursor-pointer"
            >
              <Plus className="w-4 h-4" />
              <span>Add Stream</span>
            </button>
          </form>

          {error && <p className="text-xs text-rose-400 mt-2">{error}</p>}
        </div>

        {/* Multi-Stream Grid */}
        <StreamGrid
          streams={streams}
          layout={layout}
          onLayoutChange={setLayout}
          onDeleteStream={handleDelete}
        />
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-900 py-4 text-center text-xs text-slate-500">
        StreamForge • Live RTSP Multi-Camera Viewer
      </footer>
    </div>
  );
};

export default App;
