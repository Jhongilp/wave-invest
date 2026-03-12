import type { Ticker } from '../../types';

interface WatchlistProps {
  tickers: Ticker[];
  loading: boolean;
  error: Error | null;
  onAnalyze: (ticker: string) => void;
  selectedTicker?: string;
}

export function Watchlist({ tickers, loading, error, onAnalyze, selectedTicker }: WatchlistProps) {
  if (loading) {
    return (
      <div className="p-4">
        <div className="animate-pulse space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-16 bg-gray-800 rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 text-red-400">
        <p>Error loading watchlist: {error.message}</p>
      </div>
    );
  }

  return (
    <div className="p-4 space-y-2">
      <h2 className="text-lg font-semibold text-gray-300 mb-4">Watchlist</h2>
      {tickers.map((ticker) => (
        <div
          key={ticker.symbol}
          className={`p-4 rounded-lg border transition-colors cursor-pointer ${
            selectedTicker === ticker.symbol
              ? 'bg-blue-900/30 border-blue-500'
              : 'bg-gray-800/50 border-gray-700 hover:border-gray-600'
          }`}
          onClick={() => onAnalyze(ticker.symbol)}
        >
          <div className="flex items-center justify-between">
            <div>
              <div className="font-semibold text-white">{ticker.symbol}</div>
              <div className="text-sm text-gray-400 truncate max-w-[120px]">{ticker.name}</div>
            </div>
            <div className="text-right">
              <div className="font-medium text-white">${ticker.price.toFixed(2)}</div>
              <div
                className={`text-sm ${
                  ticker.changePercent >= 0 ? 'text-green-400' : 'text-red-400'
                }`}
              >
                {ticker.changePercent >= 0 ? '+' : ''}{ticker.changePercent.toFixed(2)}%
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
