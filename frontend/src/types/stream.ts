export type StreamStatus = 'connecting' | 'live' | 'error' | 'stopped';

export interface Stream {
  id: string;
  name: string;
  rtsp_url: string;
  status: StreamStatus;
  error_message?: string;
  viewer_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateStreamPayload {
  name: string;
  rtsp_url: string;
}

export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

// Global declaration for JSMpeg loaded from public/jsmpeg.min.js
declare global {
  interface Window {
    JSMpeg: {
      Player: new (
        url: string,
        options: {
          canvas: HTMLCanvasElement;
          autoplay?: boolean;
          audio?: boolean;
          loop?: boolean;
          pauseWhenHidden?: boolean;
          disableGl?: boolean;
          disableWebAssembly?: boolean;
          onPlay?: () => void;
          onPause?: () => void;
          onEnded?: () => void;
          onStalled?: () => void;
          onSourceEstablished?: () => void;
          onSourceCompleted?: () => void;
        }
      ) => {
        play: () => void;
        pause: () => void;
        stop: () => void;
        destroy: () => void;
        volume: number;
        isPlaying: boolean;
      };
    };
  }
}
