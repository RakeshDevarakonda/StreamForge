import React from 'react';
import { LayoutGrid, Grid2X2, Grid3X3, Square, VideoOff } from 'lucide-react';
import type { CameraStream } from '../types/stream';
import { StreamPlayer } from './StreamPlayer';

export type GridLayout = '1' | '2' | '3' | 'auto';

interface StreamGridProps {
  streams: CameraStream[];
  layout: GridLayout;
  onLayoutChange: (layout: GridLayout) => void;
  onDeleteStream: (id: string) => void;
}

export const StreamGrid: React.FC<StreamGridProps> = ({
  streams,
  layout,
  onLayoutChange,
  onDeleteStream,
}) => {
  const getGridClasses = () => {
    switch (layout) {
      case '1':
        return 'grid-cols-1 max-w-4xl mx-auto';
      case '2':
        return 'grid-cols-1 sm:grid-cols-2';
      case '3':
        return 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3';
      case 'auto':
      default:
        if (streams.length <= 1) return 'grid-cols-1 max-w-4xl mx-auto';
        if (streams.length === 2) return 'grid-cols-1 sm:grid-cols-2';
        if (streams.length <= 4) return 'grid-cols-1 sm:grid-cols-2 xl:grid-cols-2';
        return 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4';
    }
  };

  return (
    <div className="space-y-4">
      {/* Header Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-2 gap-2">
        <div className="flex items-center space-x-2">
          <span className="text-sm font-semibold text-slate-300">Live Streams</span>
          <span className="px-2 py-0.5 text-xs font-bold rounded-full bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
            {streams.length}
          </span>
        </div>

        {/* Responsive Layout Switcher */}
        <div className="flex items-center space-x-1 bg-slate-900 border border-slate-800 rounded-lg p-1 self-start sm:self-auto">
          <button
            onClick={() => onLayoutChange('auto')}
            className={`px-2 py-1 text-xs rounded transition-colors cursor-pointer flex items-center space-x-1 ${
              layout === 'auto' ? 'bg-cyan-500 text-slate-950 font-bold' : 'text-slate-400 hover:text-white'
            }`}
            title="Auto Grid"
          >
            <LayoutGrid className="w-3.5 h-3.5" />
            <span className="hidden sm:inline">Auto</span>
          </button>
          <button
            onClick={() => onLayoutChange('1')}
            className={`p-1.5 rounded transition-colors cursor-pointer ${
              layout === '1' ? 'bg-cyan-500 text-slate-950' : 'text-slate-400 hover:text-white'
            }`}
            title="Single Column"
          >
            <Square className="w-3.5 h-3.5" />
          </button>
          <button
            onClick={() => onLayoutChange('2')}
            className={`p-1.5 rounded transition-colors cursor-pointer ${
              layout === '2' ? 'bg-cyan-500 text-slate-950' : 'text-slate-400 hover:text-white'
            }`}
            title="2 Columns (2x2)"
          >
            <Grid2X2 className="w-3.5 h-3.5" />
          </button>
          <button
            onClick={() => onLayoutChange('3')}
            className={`p-1.5 rounded transition-colors cursor-pointer ${
              layout === '3' ? 'bg-cyan-500 text-slate-950' : 'text-slate-400 hover:text-white'
            }`}
            title="3 Columns (3x3)"
          >
            <Grid3X3 className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {/* Grid or Empty */}
      {streams.length === 0 ? (
        <div className="border border-dashed border-slate-800 rounded-xl p-12 text-center flex flex-col items-center justify-center">
          <VideoOff className="w-8 h-8 text-slate-600 mb-2" />
          <p className="text-xs text-slate-500">No active streams. Add an RTSP stream URL above to view.</p>
        </div>
      ) : (
        <div className={`grid gap-4 ${getGridClasses()}`}>
          {streams.map((stream) => (
            <StreamPlayer key={stream.id} stream={stream} onDelete={onDeleteStream} />
          ))}
        </div>
      )}
    </div>
  );
};
