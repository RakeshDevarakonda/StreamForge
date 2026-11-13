import React from 'react';
import { LayoutGrid, Grid2X2, Grid3X3, Square, VideoOff } from 'lucide-react';
import type { Stream } from '../types/stream';
import { StreamPlayer } from './StreamPlayer';

export type GridLayout = '1' | '2' | '3' | 'auto';

interface StreamGridProps {
  streams: Stream[];
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
        return 'grid-cols-1 md:grid-cols-2';
      case '3':
        return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3';
      case 'auto':
      default:
        if (streams.length <= 1) return 'grid-cols-1 max-w-4xl mx-auto';
        if (streams.length <= 4) return 'grid-cols-1 md:grid-cols-2';
        return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4';
    }
  };

  return (
    <div className="space-y-4">
      {/* Grid Header & Layout Switcher */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-2 gap-3">
        <div className="flex items-center space-x-2.5">
          <span className="text-sm font-semibold text-slate-200">Active Live Feeds</span>
          <span className="px-2 py-0.5 text-xs font-bold rounded-full bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
            {streams.length} {streams.length === 1 ? 'Stream' : 'Streams'}
          </span>
        </div>

        {/* Layout Selectors */}
        <div className="flex items-center space-x-1 bg-slate-900 border border-slate-800 rounded-xl p-1 self-start sm:self-auto">
          <button
            onClick={() => onLayoutChange('auto')}
            className={`px-2.5 py-1 text-xs rounded-lg font-medium transition-all cursor-pointer flex items-center space-x-1.5 ${
              layout === 'auto'
                ? 'bg-cyan-500 text-slate-950 shadow-md font-semibold'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
            }`}
            title="Auto Grid Layout"
          >
            <LayoutGrid className="w-3.5 h-3.5" />
            <span>Auto</span>
          </button>
          <button
            onClick={() => onLayoutChange('1')}
            className={`p-1.5 rounded-lg transition-all cursor-pointer ${
              layout === '1'
                ? 'bg-cyan-500 text-slate-950 shadow-md'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
            }`}
            title="1 Column (Single View)"
          >
            <Square className="w-4 h-4" />
          </button>
          <button
            onClick={() => onLayoutChange('2')}
            className={`p-1.5 rounded-lg transition-all cursor-pointer ${
              layout === '2'
                ? 'bg-cyan-500 text-slate-950 shadow-md'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
            }`}
            title="2 Columns (2x2 Grid)"
          >
            <Grid2X2 className="w-4 h-4" />
          </button>
          <button
            onClick={() => onLayoutChange('3')}
            className={`p-1.5 rounded-lg transition-all cursor-pointer ${
              layout === '3'
                ? 'bg-cyan-500 text-slate-950 shadow-md'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
            }`}
            title="3 Columns (3x3 Grid)"
          >
            <Grid3X3 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Streams Grid or Empty State */}
      {streams.length === 0 ? (
        <div className="bg-slate-900/50 border border-dashed border-slate-800 rounded-2xl p-12 text-center flex flex-col items-center justify-center">
          <div className="p-4 rounded-full bg-slate-800/50 text-slate-500 mb-3">
            <VideoOff className="w-8 h-8" />
          </div>
          <h3 className="text-base font-semibold text-slate-300">No Streams Configured</h3>
          <p className="text-xs text-slate-500 max-w-sm mt-1">
            Add an RTSP stream URL using the form above to start viewing live camera feeds simultaneously.
          </p>
        </div>
      ) : (
        <div className={`grid gap-4 transition-all duration-300 ${getGridClasses()}`}>
          {streams.map((stream) => (
            <StreamPlayer
              key={stream.id}
              stream={stream}
              onDelete={onDeleteStream}
            />
          ))}
        </div>
      )}
    </div>
  );
};
