interface ConnectionStatusProps {
  isConnected: boolean;
  error: string | null;
  lastUpdate: Date | null;
  onRetry: () => void;
}

export function ConnectionStatus({ isConnected, error, lastUpdate, onRetry }: ConnectionStatusProps) {
  return (
    <div className="flex items-center justify-between mb-4 px-1">
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2">
          <div
            className={`w-2.5 h-2.5 rounded-full ${
              isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'
            }`}
          />
          <span className={`text-sm font-medium ${isConnected ? 'text-green-400' : 'text-red-400'}`}>
            {isConnected ? 'Live' : 'Disconnected'}
          </span>
        </div>
        {lastUpdate && (
          <span className="text-xs text-gray-500">
            Last update: {lastUpdate.toLocaleTimeString()}
          </span>
        )}
      </div>

      {error && (
        <div className="flex items-center gap-3">
          <span className="text-sm text-amber-400">{error}</span>
          {error.includes('Max reconnection') && (
            <button
              onClick={onRetry}
              className="px-3 py-1 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors"
            >
              Retry
            </button>
          )}
        </div>
      )}
    </div>
  );
}
