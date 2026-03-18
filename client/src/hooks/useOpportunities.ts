import { useState, useCallback } from 'react';
import { fetchApi } from '../lib/api';
import type { DailyAnalysisResult, ScoredOpportunity, ExecutionResult } from '../types';

export function useOpportunities() {
  const [opportunities, setOpportunities] = useState<ScoredOpportunity[]>([]);
  const [analysisResult, setAnalysisResult] = useState<DailyAnalysisResult | null>(null);
  const [executionResult, setExecutionResult] = useState<ExecutionResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [executing, setExecuting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const runDailyAnalysis = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetchApi<DailyAnalysisResult>('/api/daily-analysis', {
        method: 'POST',
      });
      setAnalysisResult(result);
      setOpportunities(result.opportunities);
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
      return data;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch opportunities';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
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
    executing,
    error,
    runDailyAnalysis,
    fetchOpportunities,
    executeTrades,
  };
}
