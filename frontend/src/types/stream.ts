export interface CameraStream {
  id: string;
  name: string;
  rtsp_url: string;
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
