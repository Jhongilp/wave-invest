import { useState, useCallback, useEffect } from 'react';
import { fetchApi } from '../lib/api';

export type TradingMode = 'demo' | 'real';

interface TradingModeResponse {
  mode: TradingMode;
  isDemo: boolean;
}

export function useTradingMode() {
  const [mode, setMode] = useState<TradingMode>('demo');
  const [loading, setLoading] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchMode = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchApi<TradingModeResponse>('/api/trading-mode');
      setMode(data.mode);
      return data;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch trading mode';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const updateMode = useCallback(async (newMode: TradingMode) => {
    setUpdating(true);
    setError(null);
    try {
      const data = await fetchApi<TradingModeResponse>('/api/trading-mode', {
        method: 'PUT',
        body: JSON.stringify({ mode: newMode }),
      });
      setMode(data.mode);
      return data;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update trading mode';
      setError(message);
      throw err;
    } finally {
      setUpdating(false);
    }
  }, []);

  const toggleMode = useCallback(async () => {
    const newMode = mode === 'demo' ? 'real' : 'demo';
    return updateMode(newMode);
  }, [mode, updateMode]);

  const isDemo = mode === 'demo';

  // Fetch mode on mount
  useEffect(() => {
    fetchMode().catch(() => {});
  }, [fetchMode]);

  return {
    mode,
    isDemo,
    loading,
    updating,
    error,
    fetchMode,
    updateMode,
    toggleMode,
  };
}
