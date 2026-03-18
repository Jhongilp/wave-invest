import { useState, useCallback, useEffect } from 'react';
import { fetchApi } from '../lib/api';
import type { PortfolioSummary, Position, Transaction } from '../types';

export function usePortfolio() {
  const [portfolio, setPortfolio] = useState<PortfolioSummary | null>(null);
  const [positions, setPositions] = useState<Position[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchPortfolio = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchApi<PortfolioSummary>('/api/portfolio');
      setPortfolio(data);
      setPositions(data.openPositions || []);
      return data;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch portfolio';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const createPortfolio = useCallback(async (budget: number) => {
    setLoading(true);
    setError(null);
    try {
      await fetchApi('/api/portfolio', {
        method: 'POST',
        body: JSON.stringify({ budget }),
      });
      return fetchPortfolio();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create portfolio';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [fetchPortfolio]);

  const fetchPositions = useCallback(async () => {
    try {
      const data = await fetchApi<Position[]>('/api/positions');
      setPositions(data);
      return data;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch positions';
      setError(message);
      throw err;
    }
  }, []);

  const closePosition = useCallback(async (positionId: string) => {
    try {
      await fetchApi(`/api/positions/${positionId}`, {
        method: 'DELETE',
      });
      await fetchPositions();
      await fetchPortfolio();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to close position';
      setError(message);
      throw err;
    }
  }, [fetchPositions, fetchPortfolio]);

  const fetchTransactions = useCallback(async (since?: string) => {
    try {
      const endpoint = since ? `/api/transactions?since=${since}` : '/api/transactions';
      const data = await fetchApi<Transaction[]>(endpoint);
      setTransactions(data);
      return data;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch transactions';
      setError(message);
      throw err;
    }
  }, []);

  // Fetch portfolio on mount
  useEffect(() => {
    fetchPortfolio().catch(() => {});
  }, [fetchPortfolio]);

  return {
    portfolio,
    positions,
    transactions,
    loading,
    error,
    fetchPortfolio,
    createPortfolio,
    fetchPositions,
    closePosition,
    fetchTransactions,
  };
}
