import React, { useEffect, useRef, useState, useCallback } from 'react';
import { Play, Pause, RefreshCw, Trash2, Maximize2 } from 'lucide-react';
import type { CameraStream } from '../types/stream';
import { getStreamWebSocketUrl } from '../services/api';

interface StreamPlayerProps {
  stream: CameraStream;
  onDelete: (id: string) => void;
}

export const StreamPlayer: React.FC<StreamPlayerProps> = ({ stream, onDelete }) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const playerRef = useRef<any>(null);

  const [isPlaying, setIsPlaying] = useState<boolean>(true);
  const [isLive, setIsLive] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const startStream = useCallback(() => {
    if (!canvasRef.current || !window.JSMpeg) return;

    if (playerRef.current) {
      try { playerRef.current.destroy(); } catch {}
      playerRef.current = null;
    }

    setIsLive(false);
    setError(null);

    try {
      const wsUrl = getStreamWebSocketUrl(stream.rtsp_url);
      playerRef.current = new window.JSMpeg.Player(wsUrl, {
        canvas: canvasRef.current,
        autoplay: true,
        audio: false,
        onSourceEstablished: () => {
          setIsLive(true);
          setIsPlaying(true);
        },
        onSourceCompleted: () => {
          setIsLive(false);
          setIsPlaying(false);
          setError('Stream disconnected');
        },
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Connection failed');
    }
  }, [stream.rtsp_url]);

  useEffect(() => {
    startStream();
    return () => {
      if (playerRef.current) {
        try { playerRef.current.destroy(); } catch {}
        playerRef.current = null;
      }
    };
  }, [startStream]);

  const togglePlay = () => {
    if (!playerRef.current) {
      startStream();
      return;
    }
    if (isPlaying) {
      try { playerRef.current.destroy(); } catch {}
      playerRef.current = null;
      setIsPlaying(false);
      setIsLive(false);
    } else {
      startStream();
    }
  };

  const handleFullscreen = () => {
    if (!containerRef.current) return;
    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      containerRef.current.requestFullscreen();
    }
  };

  return (
    <div
      ref={containerRef}
      className="flex flex-col bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-lg"
    >
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 bg-slate-900/90 border-b border-slate-800">
        <div className="flex items-center space-x-2 truncate min-w-0 mr-2">
          <span className={`h-2.5 w-2.5 shrink-0 rounded-full ${isLive ? 'bg-emerald-500 animate-pulse' : 'bg-slate-600'}`} />
          <span className="font-medium text-xs text-slate-200 truncate">{stream.name}</span>
        </div>
        <button
          onClick={() => onDelete(stream.id)}
          title="Delete stream"
          className="text-slate-400 hover:text-rose-400 p-1 rounded transition-colors cursor-pointer shrink-0"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* Video Canvas */}
      <div className="relative aspect-video bg-black flex items-center justify-center overflow-hidden">
        <canvas ref={canvasRef} className="w-full h-full object-contain block" />

        {!isLive && !error && (
          <div className="absolute inset-0 bg-slate-950/70 flex flex-col items-center justify-center space-y-2">
            <RefreshCw className="w-6 h-6 text-cyan-400 animate-spin" />
            <span className="text-xs text-slate-300">Connecting...</span>
          </div>
        )}

        {error && (
          <div className="absolute inset-0 bg-slate-950/85 flex flex-col items-center justify-center p-3 text-center">
            <p className="text-xs text-rose-400 font-semibold">{error}</p>
            <button
              onClick={startStream}
              className="mt-2 px-3 py-1 bg-slate-800 hover:bg-slate-700 text-xs rounded text-slate-200 cursor-pointer"
            >
              Retry
            </button>
          </div>
        )}

        {/* Controls Overlay */}
        <div className="absolute bottom-2 left-2 right-2 flex items-center justify-between opacity-0 hover:opacity-100 transition-opacity">
          <div className="flex space-x-1">
            <button
              onClick={togglePlay}
              className="p-1.5 bg-slate-950/80 hover:bg-slate-800 text-white rounded cursor-pointer"
            >
              {isPlaying ? <Pause className="w-3.5 h-3.5" /> : <Play className="w-3.5 h-3.5" />}
            </button>
            <button
              onClick={startStream}
              className="p-1.5 bg-slate-950/80 hover:bg-slate-800 text-white rounded cursor-pointer"
            >
              <RefreshCw className="w-3.5 h-3.5" />
            </button>
          </div>
          <button
            onClick={handleFullscreen}
            className="p-1.5 bg-slate-950/80 hover:bg-slate-800 text-white rounded cursor-pointer"
          >
            <Maximize2 className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {/* URL Footer */}
      <div className="px-3 py-1.5 bg-slate-950 text-[10px] text-slate-500 font-mono truncate" title={stream.rtsp_url}>
        {stream.rtsp_url}
      </div>
    </div>
  );
};
