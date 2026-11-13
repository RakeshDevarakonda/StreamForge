import { useState, useEffect } from 'react';
import { Radio, PlusCircle, Video, AlertCircle } from 'lucide-react';
import type { Stream, CreateStreamPayload } from './types/stream';
import { getSavedCameras, saveCameraToStorage, removeCameraFromStorage } from './services/storage';
import { StreamGrid } from './components/StreamGrid';
import type { GridLayout } from './components/StreamGrid';

export const App = () => {
  const [streams, setStreams] = useState<Stream[]>([]);
  const [layout, setLayout] = useState<GridLayout>('auto');
  const [name, setName] = useState('');
  const [rtspUrl, setRtspUrl] = useState('');
  const [error, setError] = useState<string | null>(null);

  // Load cameras strictly from frontend localStorage
  useEffect(() => {
    const saved = getSavedCameras();
    const formatted: Stream[] = saved.map((cam, idx) => ({
      id: `cam-${idx}-${Date.now().toString().slice(-4)}`,
      name: cam.name,
      rtsp_url: cam.rtsp_url,
      status: 'live',
      viewer_count: 1,
      created_at: cam.saved_at,
      updated_at: cam.saved_at,
    }));
    setStreams(formatted);
  }, []);

  const handleAddStream = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const trimmedUrl = rtspUrl.trim();
    if (!trimmedUrl) {
      setError('Please provide an RTSP stream URL');
      return;
    }

    if (!trimmedUrl.startsWith('rtsp://') && !trimmedUrl.startsWith('rtsps://')) {
      setError('Invalid URL format: Must start with rtsp:// or rtsps://');
      return;
    }

    const payload: CreateStreamPayload = {
      name: name.trim() || `Camera ${streams.length + 1}`,
      rtsp_url: trimmedUrl,
    };

    // Save to localStorage
    saveCameraToStorage(payload);

    const newStream: Stream = {
      id: `cam-${Date.now()}`,
      name: payload.name,
      rtsp_url: payload.rtsp_url,
      status: 'live',
      viewer_count: 1,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    setStreams((prev) => [...prev, newStream]);
    setName('');
    setRtspUrl('');
  };

  const handleDeleteStream = (id: string) => {
    const target = streams.find((s) => s.id === id);
    if (target) {
      removeCameraFromStorage(target.rtsp_url);
    }
    setStreams((prev) => prev.filter((s) => s.id !== id));
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      {/* Top Header */}
      <header className="sticky top-0 z-50 bg-slate-950/80 backdrop-blur-md border-b border-slate-800/80 px-4 sm:px-8 py-3.5">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="h-9 w-9 rounded-xl bg-gradient-to-tr from-cyan-500 to-blue-600 flex items-center justify-center shadow-lg shadow-cyan-500/20">
              <Radio className="w-5 h-5 text-slate-950" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h1 className="font-bold text-base tracking-tight text-white">StreamForge</h1>
                <span className="text-[10px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-400 border border-cyan-500/30">
                  RTSP Viewer
                </span>
              </div>
              <p className="text-[11px] text-slate-400">On-demand RTSP live video streaming over WebSockets</p>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8 space-y-8">
        {/* Input Form Section */}
        <section className="bg-slate-900/90 border border-slate-800 rounded-2xl p-5 shadow-2xl backdrop-blur-sm">
          <div className="pb-4 mb-4 border-b border-slate-800/80">
            <div className="flex items-center space-x-3">
              <div className="p-2.5 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">
                <Video className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-base font-semibold text-slate-100">Add RTSP Stream</h2>
                <p className="text-xs text-slate-400">Enter your live RTSP camera feed URL</p>
              </div>
            </div>
          </div>

          <form onSubmit={handleAddStream} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-12 gap-3">
              <div className="md:col-span-4">
                <label className="block text-xs font-medium text-slate-300 mb-1.5">
                  Camera Name <span className="text-slate-500 font-normal">(Optional)</span>
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Front Gate, Lobby"
                  className="w-full px-3.5 py-2.5 bg-slate-950/70 border border-slate-800 rounded-xl text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500/80 focus:ring-1 focus:ring-cyan-500/50 transition-all"
                />
              </div>

              <div className="md:col-span-8">
                <label className="block text-xs font-medium text-slate-300 mb-1.5">
                  RTSP Stream URL <span className="text-rose-400">*</span>
                </label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={rtspUrl}
                    onChange={(e) => setRtspUrl(e.target.value)}
                    placeholder="rtsp://your-camera-ip:554/stream or rtsp://localhost:8554/..."
                    className="flex-1 px-3.5 py-2.5 bg-slate-950/70 border border-slate-800 rounded-xl text-sm font-mono text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500/80 focus:ring-1 focus:ring-cyan-500/50 transition-all"
                  />
                  <button
                    type="submit"
                    className="px-5 py-2.5 bg-cyan-500 hover:bg-cyan-400 active:bg-cyan-600 text-slate-950 font-semibold text-sm rounded-xl transition-all shadow-lg shadow-cyan-500/20 flex items-center justify-center space-x-2 shrink-0 cursor-pointer"
                  >
                    <PlusCircle className="w-4 h-4" />
                    <span>Add Stream</span>
                  </button>
                </div>
              </div>
            </div>

            {error && (
              <div className="p-3 bg-rose-500/10 border border-rose-500/20 rounded-xl flex items-center space-x-2 text-rose-300 text-xs">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{error}</span>
              </div>
            )}
          </form>
        </section>

        {/* Live Grid Stream Section */}
        <section>
          <StreamGrid
            streams={streams}
            layout={layout}
            onLayoutChange={setLayout}
            onDeleteStream={handleDeleteStream}
          />
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-900 py-6 px-4 text-center text-xs text-slate-500">
        <p>StreamForge • Ultra Low-Latency Multi-Camera RTSP Live Streaming Platform</p>
      </footer>
    </div>
  );
};

export default App;
