import React, { useState, useEffect, useCallback } from 'react';
import { Radio, RefreshCw, Cpu, Layers } from 'lucide-react';
import type { Stream, CreateStreamPayload } from './types/stream';
import { fetchStreams, createStream, deleteStream } from './services/api';
import { getSavedCameras, saveCameraToStorage, removeCameraFromStorage } from './services/storage';
import { AddStreamForm } from './components/AddStreamForm';
import { StreamGrid } from './components/StreamGrid';
import type { GridLayout } from './components/StreamGrid';

export const App: React.FC = () => {
  const [streams, setStreams] = useState<Stream[]>([]);
  const [layout, setLayout] = useState<GridLayout>('auto');
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [fetchError, setFetchError] = useState<string | null>(null);

  const loadStreams = useCallback(async () => {
    try {
      setFetchError(null);
      let backendStreams = await fetchStreams();

      // Auto-rehydrate from localStorage if backend was restarted or empty
      const savedCameras = getSavedCameras();
      if (savedCameras.length > 0) {
        const existingUrls = new Set(backendStreams.map((s) => s.rtsp_url));
        const missingFromBackend = savedCameras.filter((c) => !existingUrls.has(c.rtsp_url));

        if (missingFromBackend.length > 0) {
          // Re-register saved cameras in backend so FFmpeg can stream them
          for (const cam of missingFromBackend) {
            try {
              const registered = await createStream({
                name: cam.name,
                rtsp_url: cam.rtsp_url,
              });
              backendStreams.push(registered);
            } catch {
              // Ignore conflicts if already being registered
            }
          }
        }
      }

      setStreams(backendStreams);
    } catch (err: unknown) {
      setFetchError(err instanceof Error ? err.message : 'Failed to connect to backend server');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadStreams();
    // Poll backend every 5 seconds to update viewer counts and stream status
    const interval = setInterval(async () => {
      try {
        const data = await fetchStreams();
        setStreams(data);
      } catch {
        // Silent polling error handling
      }
    }, 5000);
    return () => clearInterval(interval);
  }, [loadStreams]);

  const handleAddStream = async (payload: CreateStreamPayload) => {
    setIsSubmitting(true);
    try {
      const newStream = await createStream(payload);
      // Persist in browser localStorage
      saveCameraToStorage(payload);
      setStreams((prev) => [...prev, newStream]);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteStream = async (id: string) => {
    try {
      const target = streams.find((s) => s.id === id);
      await deleteStream(id);
      if (target) {
        // Remove from browser localStorage
        removeCameraFromStorage(target.rtsp_url);
      }
      setStreams((prev) => prev.filter((s) => s.id !== id));
    } catch (err) {
      console.error('Delete stream error:', err);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      {/* Top Navigation Bar */}
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
              <p className="text-[11px] text-slate-400">Low-latency RTSP live video streaming over WebSockets</p>
            </div>
          </div>

          <div className="flex items-center space-x-3">
            <div className="hidden md:flex items-center space-x-4 text-xs text-slate-400">
              <span className="flex items-center space-x-1.5">
                <Cpu className="w-3.5 h-3.5 text-cyan-400" />
                <span>Go & FFmpeg</span>
              </span>
              <span className="flex items-center space-x-1.5">
                <Layers className="w-3.5 h-3.5 text-emerald-400" />
                <span>WebSocket / MPEG-TS</span>
              </span>
            </div>

            <button
              onClick={loadStreams}
              disabled={isLoading}
              title="Refresh streams"
              className="p-2 rounded-xl bg-slate-900 border border-slate-800 text-slate-300 hover:text-white hover:bg-slate-800 transition-colors cursor-pointer"
            >
              <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8 space-y-8">
        {/* Connection Error Banner */}
        {fetchError && (
          <div className="p-4 bg-rose-500/10 border border-rose-500/30 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-rose-300">
            <div>
              <p className="text-sm font-semibold">Backend Connection Failed</p>
              <p className="text-xs text-rose-400/80 mt-0.5">
                Unable to reach Go server at http://localhost:8080. Ensure the backend server is running.
              </p>
            </div>
            <button
              onClick={loadStreams}
              className="px-3.5 py-1.5 bg-rose-500/20 hover:bg-rose-500/30 text-rose-200 text-xs font-medium rounded-xl transition-colors cursor-pointer shrink-0 self-start sm:self-auto"
            >
              Retry Connection
            </button>
          </div>
        )}

        {/* Input Form Section */}
        <section>
          <AddStreamForm
            onAddStream={handleAddStream}
            isLoading={isSubmitting}
          />
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
        <p>Built for Skylark Labs Full Stack Engineer Assessment • React, TypeScript, Go & FFmpeg</p>
      </footer>
    </div>
  );
};

export default App;
