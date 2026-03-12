import { useState, useEffect, useCallback } from 'react';
import type { Ticker, WatchlistResponse } from '../types';
import { fetchApi } from '../lib/api';

interface UseWatchlistResult {
  tickers: Ticker[];
  loading: boolean;
  error: Error | null;
  refetch: () => void;
}

export function useWatchlist(): UseWatchlistResult {
  const [tickers, setTickers] = useState<Ticker[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchWatchlist = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchApi<WatchlistResponse>('/api/watchlist');
      setTickers(data.tickers);
    } catch (e) {
      setError(e instanceof Error ? e : new Error('Failed to fetch watchlist'));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchWatchlist();
  }, [fetchWatchlist]);

  return { tickers, loading, error, refetch: fetchWatchlist };
}
