import React, { useEffect, useRef, useState, useCallback } from 'react';
import { Play, Pause, RefreshCw, Trash2, Maximize2, AlertCircle, Radio } from 'lucide-react';
import type { Stream } from '../types/stream';
import { getStreamWebSocketUrl } from '../services/api';

interface StreamPlayerProps {
  stream: Stream;
  onDelete: (id: string) => void;
}

export const StreamPlayer: React.FC<StreamPlayerProps> = ({ stream, onDelete }) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const playerRef = useRef<any>(null);

  const [isPlaying, setIsPlaying] = useState<boolean>(false);
  const [isLive, setIsLive] = useState<boolean>(false);
  const [isConnecting, setIsConnecting] = useState<boolean>(true);
  const [playerError, setPlayerError] = useState<string | null>(null);

  const initPlayer = useCallback(() => {
    if (!canvasRef.current) return;

    if (!window.JSMpeg) {
      setIsConnecting(false);
      setPlayerError('JSMpeg player library not available');
      return;
    }

    // Cleanup previous player instance
    if (playerRef.current) {
      try {
        playerRef.current.destroy();
      } catch {
        // ignore
      }
      playerRef.current = null;
    }

    setIsConnecting(true);
    setIsPlaying(false);
    setIsLive(false);
    setPlayerError(null);

    const wsUrl = getStreamWebSocketUrl(stream.id);

    try {
      playerRef.current = new window.JSMpeg.Player(wsUrl, {
        canvas: canvasRef.current,
        autoplay: true,
        audio: false,
        pauseWhenHidden: false,
        disableGl: false,
        onSourceEstablished: () => {
          setIsConnecting(false);
          setIsPlaying(true);
          setIsLive(true);
          setPlayerError(null);
        },
        onSourceCompleted: () => {
          setIsConnecting(false);
          setIsLive(false);
          setIsPlaying(false);
          setPlayerError('Stream closed by server');
        },
        onStalled: () => {
          setIsConnecting(true);
          setIsLive(false);
          // Set timeout to avoid endless spinner if connection is permanently stalled
          setTimeout(() => {
            if (!playerRef.current?.isPlaying) {
              setIsConnecting(false);
              setPlayerError('Stream stalled: Reconnection timeout');
            }
          }, 8000);
        },
      });
    } catch (err: unknown) {
      setIsConnecting(false);
      setIsPlaying(false);
      setIsLive(false);
      setPlayerError(err instanceof Error ? err.message : 'Failed to initialize player');
    }
  }, [stream.id]);

  useEffect(() => {
    initPlayer();

    return () => {
      if (playerRef.current) {
        try {
          playerRef.current.destroy();
        } catch {
          // ignore
        }
        playerRef.current = null;
      }
    };
  }, [initPlayer]);

  const togglePlay = () => {
    if (!playerRef.current) {
      initPlayer();
      return;
    }

    if (isPlaying) {
      // Cleanly pause and destroy current WebSocket connection so backend stops streaming
      try {
        playerRef.current.destroy();
      } catch {
        // ignore
      }
      playerRef.current = null;
      setIsPlaying(false);
      setIsLive(false);
    } else {
      initPlayer();
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

  // Helper to safely mask credentials in RTSP URLs (e.g., rtsp://user:pass@host -> rtsp://***:***@host)
  const sanitizeUrl = (url: string) => {
    try {
      const parsed = new URL(url);
      if (parsed.username || parsed.password) {
        parsed.username = '***';
        parsed.password = '***';
        return parsed.toString();
      }
      return url;
    } catch {
      return url.replace(/\/\/([^:]+):([^@]+)@/, '//***:***@');
    }
  };

  return (
    <div
      ref={containerRef}
      className="group relative flex flex-col bg-slate-900/90 border border-slate-800 rounded-xl overflow-hidden shadow-xl transition-all duration-300 hover:border-slate-700 hover:shadow-cyan-950/20"
    >
      {/* Stream Card Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-slate-900/80 border-b border-slate-800">
        <div className="flex items-center space-x-2.5 overflow-hidden">
          <div className="relative flex items-center justify-center">
            {isLive && !playerError ? (
              <span className="flex h-3 w-3">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
              </span>
            ) : (
              <span className="h-3 w-3 rounded-full bg-slate-600" />
            )}
          </div>
          <span className="font-semibold text-sm text-slate-200 truncate">{stream.name}</span>
        </div>

        <div className="flex items-center space-x-2">
          <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-slate-800/80 text-slate-400 border border-slate-700/50">
            {stream.viewer_count > 0 ? `${stream.viewer_count} viewer${stream.viewer_count > 1 ? 's' : ''}` : 'Ready'}
          </span>
          <button
            onClick={() => onDelete(stream.id)}
            title="Remove stream"
            className="p-1.5 text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition-colors cursor-pointer"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Video Canvas Container */}
      <div className="relative aspect-video w-full bg-black flex items-center justify-center overflow-hidden">
        <canvas
          ref={canvasRef}
          className="w-full h-full object-contain block"
        />

        {/* Connecting / Loading Spinner */}
        {isConnecting && !playerError && (
          <div className="absolute inset-0 bg-slate-950/70 backdrop-blur-xs flex flex-col items-center justify-center space-y-3 z-10">
            <RefreshCw className="w-8 h-8 text-cyan-400 animate-spin" />
            <p className="text-xs font-medium text-slate-300 tracking-wide">Connecting to RTSP feed...</p>
          </div>
        )}

        {/* Error Overlay */}
        {playerError && (
          <div className="absolute inset-0 bg-slate-950/85 backdrop-blur-xs flex flex-col items-center justify-center p-4 text-center z-10">
            <AlertCircle className="w-8 h-8 text-rose-500 mb-2" />
            <p className="text-sm font-semibold text-rose-300">Stream Unavailable</p>
            <p className="text-xs text-slate-400 max-w-xs mt-1 truncate">{playerError}</p>
            <button
              onClick={initPlayer}
              className="mt-3 px-3 py-1.5 text-xs font-medium bg-rose-500/20 hover:bg-rose-500/30 text-rose-300 rounded-md transition-colors cursor-pointer flex items-center space-x-1.5"
            >
              <RefreshCw className="w-3.5 h-3.5" />
              <span>Retry</span>
            </button>
          </div>
        )}

        {/* Bottom Control Overlay Bar (Visible on Hover) */}
        <div className="absolute bottom-0 inset-x-0 p-3 bg-gradient-to-t from-slate-950/90 via-slate-950/50 to-transparent flex items-center justify-between opacity-0 group-hover:opacity-100 transition-opacity duration-200 z-20">
          <div className="flex items-center space-x-2">
            <button
              onClick={togglePlay}
              className="p-1.5 bg-slate-800/80 hover:bg-cyan-500 text-slate-200 hover:text-white rounded-lg transition-colors cursor-pointer shadow-sm"
              title={isPlaying ? 'Pause' : 'Play'}
            >
              {isPlaying ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
            </button>
            <button
              onClick={initPlayer}
              className="p-1.5 bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white rounded-lg transition-colors cursor-pointer"
              title="Reconnect"
            >
              <RefreshCw className="w-4 h-4" />
            </button>
          </div>

          {isLive && (
            <div className="flex items-center space-x-2">
              <span className="text-[10px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 flex items-center space-x-1">
                <Radio className="w-3 h-3 animate-pulse" />
                <span>LIVE</span>
              </span>
              <button
                onClick={handleFullscreen}
                className="p-1.5 bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white rounded-lg transition-colors cursor-pointer"
                title="Fullscreen"
              >
                <Maximize2 className="w-4 h-4" />
              </button>
            </div>
          )}
        </div>
      </div>

      {/* URL Footer */}
      <div className="px-3.5 py-2 bg-slate-950/60 border-t border-slate-800/50 flex items-center justify-between text-[11px] text-slate-400">
        <span className="truncate max-w-[280px] font-mono text-slate-500" title={sanitizeUrl(stream.rtsp_url)}>
          {sanitizeUrl(stream.rtsp_url)}
        </span>
        <span className="shrink-0 text-slate-500">MPEG-TS / WS</span>
      </div>
    </div>
  );
};
