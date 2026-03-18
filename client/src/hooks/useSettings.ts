import { useState, useCallback, useEffect } from 'react';
import { fetchApi } from '../lib/api';
import type { TradingRules } from '../types';
import { DEFAULT_TRADING_RULES } from '../types';

export function useSettings() {
  const [settings, setSettings] = useState<TradingRules>(DEFAULT_TRADING_RULES);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSettings = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchApi<TradingRules>('/api/settings');
      setSettings(data);
      return data;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch settings';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const updateSettings = useCallback(async (newSettings: TradingRules) => {
    setSaving(true);
    setError(null);
    try {
      const data = await fetchApi<TradingRules>('/api/settings', {
        method: 'PUT',
        body: JSON.stringify(newSettings),
      });
      setSettings(data);
      return data;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update settings';
      setError(message);
      throw err;
    } finally {
      setSaving(false);
    }
  }, []);

  const resetToDefaults = useCallback(() => {
    setSettings(DEFAULT_TRADING_RULES);
  }, []);

  // Fetch settings on mount
  useEffect(() => {
    fetchSettings().catch(() => {});
  }, [fetchSettings]);

  return {
    settings,
    loading,
    saving,
    error,
    setSettings,
    fetchSettings,
    updateSettings,
    resetToDefaults,
  };
}
