import { useState, useCallback } from 'react';
import type { TradingPlan } from '../types';
import { fetchApi } from '../lib/api';

interface UseTradingPlanResult {
  plan: TradingPlan | null;
  loading: boolean;
  error: Error | null;
  analyze: (ticker: string) => Promise<void>;
  fetchCached: (ticker: string) => Promise<void>;
}

export function useTradingPlan(): UseTradingPlanResult {
  const [plan, setPlan] = useState<TradingPlan | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const analyze = useCallback(async (ticker: string) => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchApi<TradingPlan>(`/api/analyze/${ticker}`, {
        method: 'POST',
      });
      setPlan(data);
    } catch (e) {
      setError(e instanceof Error ? e : new Error('Failed to analyze ticker'));
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchCached = useCallback(async (ticker: string) => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchApi<TradingPlan>(`/api/plan/${ticker}`);
      setPlan(data);
    } catch (e) {
      setError(e instanceof Error ? e : new Error('Failed to fetch trading plan'));
    } finally {
      setLoading(false);
    }
  }, []);

  return { plan, loading, error, analyze, fetchCached };
}
