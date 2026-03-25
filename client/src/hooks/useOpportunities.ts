import { useState, useCallback, useEffect } from 'react';
import { fetchApi } from '../lib/api';
import type { DailyAnalysisResult, ScoredOpportunity, ExecutionResult } from '../types';

export function useOpportunities() {
  const [opportunities, setOpportunities] = useState<ScoredOpportunity[]>([]);
  const [analysisResult, setAnalysisResult] = useState<DailyAnalysisResult | null>(null);
  const [executionResult, setExecutionResult] = useState<ExecutionResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [executing, setExecuting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastAnalysisDate, setLastAnalysisDate] = useState<string | null>(null);

  const runDailyAnalysis = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetchApi<DailyAnalysisResult>('/api/daily-analysis', {
        method: 'POST',
      });
      setAnalysisResult(result);
      setOpportunities(result.opportunities);
      setLastAnalysisDate(result.date);
      return result;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to run daily analysis';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchOpportunities = useCallback(async (date?: string) => {
    setLoading(true);
    setError(null);
    try {
      const endpoint = date ? `/api/opportunities/${date}` : '/api/opportunities';
      const data = await fetchApi<ScoredOpportunity[]>(endpoint);
      setOpportunities(data);
      // Get date from first opportunity or use provided date
      if (data.length > 0) {
        const analysisDate = date || new Date().toISOString().split('T')[0];
        setLastAnalysisDate(analysisDate);
      }
      return data;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch opportunities';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Auto-fetch today's opportunities on mount
  useEffect(() => {
    const loadPersistedData = async () => {
      try {
        const data = await fetchApi<ScoredOpportunity[]>('/api/opportunities');
        setOpportunities(data);
        if (data.length > 0) {
          setLastAnalysisDate(new Date().toISOString().split('T')[0]);
        }
      } catch {
        // No persisted data for today, that's ok
      } finally {
        setInitialLoading(false);
      }
    };
    loadPersistedData();
  }, []);

  const executeTrades = useCallback(async () => {
    setExecuting(true);
    setError(null);
    try {
      const result = await fetchApi<ExecutionResult>('/api/execute-trades', {
        method: 'POST',
      });
      setExecutionResult(result);
      return result;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to execute trades';
      setError(message);
      throw err;
    } finally {
      setExecuting(false);
    }
  }, []);

  return {
    opportunities,
    analysisResult,
    executionResult,
    loading,
    initialLoading,
    executing,
    error,
    lastAnalysisDate,
    runDailyAnalysis,
    fetchOpportunities,
    executeTrades,
  };
}
