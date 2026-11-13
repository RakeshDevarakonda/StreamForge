import React, { useState } from 'react';
import { PlusCircle, Video, AlertCircle } from 'lucide-react';
import type { CreateStreamPayload } from '../types/stream';

interface AddStreamFormProps {
  onAddStream: (payload: CreateStreamPayload) => Promise<void>;
  isLoading: boolean;
}

export const AddStreamForm: React.FC<AddStreamFormProps> = ({ onAddStream, isLoading }) => {
  const [name, setName] = useState('');
  const [rtspUrl, setRtspUrl] = useState('');
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
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

    try {
      await onAddStream({
        name: name.trim() || `Camera ${Date.now().toString().slice(-4)}`,
        rtsp_url: trimmedUrl,
      });
      setName('');
      setRtspUrl('');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to add stream');
    }
  };

  return (
    <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-5 shadow-2xl backdrop-blur-sm">
      <div className="pb-4 mb-4 border-b border-slate-800/80">
        <div className="flex items-center space-x-3">
          <div className="p-2.5 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">
            <Video className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-base font-semibold text-slate-100">Add RTSP Stream</h2>
            <p className="text-xs text-slate-400">Add any live RTSP camera feed URL</p>
          </div>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
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
              disabled={isLoading}
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
                disabled={isLoading}
                className="flex-1 px-3.5 py-2.5 bg-slate-950/70 border border-slate-800 rounded-xl text-sm font-mono text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500/80 focus:ring-1 focus:ring-cyan-500/50 transition-all"
              />
              <button
                type="submit"
                disabled={isLoading}
                className="px-5 py-2.5 bg-cyan-500 hover:bg-cyan-400 active:bg-cyan-600 disabled:opacity-50 text-slate-950 font-semibold text-sm rounded-xl transition-all shadow-lg shadow-cyan-500/20 flex items-center justify-center space-x-2 shrink-0 cursor-pointer"
              >
                <PlusCircle className="w-4 h-4" />
                <span>{isLoading ? 'Adding...' : 'Add Stream'}</span>
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
    </div>
  );
};
