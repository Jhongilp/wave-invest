import { useSettings } from '../../hooks';
import type { TradingRules } from '../../types';
import { DEFAULT_TRADING_RULES } from '../../types';

export function SettingsView() {
  const {
    settings,
    loading,
    saving,
    error,
    setSettings,
    updateSettings,
    resetToDefaults,
  } = useSettings();

  const handleChange = (field: keyof TradingRules, value: string) => {
    const numValue = parseFloat(value);
    if (!isNaN(numValue)) {
      setSettings({ ...settings, [field]: numValue });
    }
  };

  const handleSave = async () => {
    await updateSettings(settings);
  };

  const handleReset = () => {
    resetToDefaults();
  };

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center">
        <p className="text-gray-400">Loading settings...</p>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Trading Settings</h1>
      </div>

      {error && (
        <div className="bg-red-900/50 border border-red-500 text-red-200 px-4 py-3 rounded-lg mb-6">
          {error}
        </div>
      )}

      <div className="max-w-2xl">
        <div className="bg-gray-800 rounded-lg p-6">
          <div className="space-y-6">
            {/* Max Position Percent */}
            <div>
              <label className="block text-gray-300 font-medium mb-2">
                Maximum Position Size (% of portfolio)
              </label>
              <div className="flex items-center gap-4">
                <input
                  type="range"
                  min="0.01"
                  max="0.25"
                  step="0.01"
                  value={settings.maxPositionPercent}
                  onChange={(e) => handleChange('maxPositionPercent', e.target.value)}
                  className="flex-1"
                />
                <span className="text-white w-16 text-right">
                  {(settings.maxPositionPercent * 100).toFixed(0)}%
                </span>
              </div>
              <p className="text-gray-500 text-sm mt-1">
                Default: {(DEFAULT_TRADING_RULES.maxPositionPercent * 100).toFixed(0)}%
              </p>
            </div>

            {/* Max Concurrent Positions */}
            <div>
              <label className="block text-gray-300 font-medium mb-2">
                Maximum Concurrent Positions
              </label>
              <div className="flex items-center gap-4">
                <input
                  type="range"
                  min="1"
                  max="20"
                  step="1"
                  value={settings.maxConcurrentPositions}
                  onChange={(e) => handleChange('maxConcurrentPositions', e.target.value)}
                  className="flex-1"
                />
                <span className="text-white w-16 text-right">
                  {settings.maxConcurrentPositions}
                </span>
              </div>
              <p className="text-gray-500 text-sm mt-1">
                Default: {DEFAULT_TRADING_RULES.maxConcurrentPositions}
              </p>
            </div>

            {/* Min Score Threshold */}
            <div>
              <label className="block text-gray-300 font-medium mb-2">
                Minimum Score Threshold
              </label>
              <div className="flex items-center gap-4">
                <input
                  type="range"
                  min="0"
                  max="100"
                  step="5"
                  value={settings.minScoreThreshold}
                  onChange={(e) => handleChange('minScoreThreshold', e.target.value)}
                  className="flex-1"
                />
                <span className="text-white w-16 text-right">
                  {settings.minScoreThreshold}
                </span>
              </div>
              <p className="text-gray-500 text-sm mt-1">
                Only execute trades with score ≥ {settings.minScoreThreshold}. Default: {DEFAULT_TRADING_RULES.minScoreThreshold}
              </p>
            </div>

            {/* Daily Loss Limit */}
            <div>
              <label className="block text-gray-300 font-medium mb-2">
                Daily Loss Limit (% of portfolio)
              </label>
              <div className="flex items-center gap-4">
                <input
                  type="range"
                  min="0.01"
                  max="0.25"
                  step="0.01"
                  value={settings.dailyLossLimit}
                  onChange={(e) => handleChange('dailyLossLimit', e.target.value)}
                  className="flex-1"
                />
                <span className="text-white w-16 text-right">
                  {(settings.dailyLossLimit * 100).toFixed(0)}%
                </span>
              </div>
              <p className="text-gray-500 text-sm mt-1">
                Stop trading if daily loss exceeds this limit. Default: {(DEFAULT_TRADING_RULES.dailyLossLimit * 100).toFixed(0)}%
              </p>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-4 mt-8 pt-6 border-t border-gray-700">
            <button
              onClick={handleSave}
              disabled={saving}
              className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 text-white px-6 py-2 rounded-lg font-medium transition-colors"
            >
              {saving ? 'Saving...' : 'Save Settings'}
            </button>
            <button
              onClick={handleReset}
              className="text-gray-400 hover:text-white px-4 py-2"
            >
              Reset to Defaults
            </button>
          </div>
        </div>

        {/* Info Box */}
        <div className="bg-gray-800/50 rounded-lg p-4 mt-6">
          <h3 className="text-gray-300 font-medium mb-2">About Trading Rules</h3>
          <ul className="text-gray-500 text-sm space-y-2">
            <li>• <strong>Position Size:</strong> Maximum % of your portfolio allocated to a single trade</li>
            <li>• <strong>Concurrent Positions:</strong> Maximum number of open positions at any time</li>
            <li>• <strong>Score Threshold:</strong> Opportunities below this score won't be executed</li>
            <li>• <strong>Daily Loss Limit:</strong> Trading stops if your daily losses exceed this threshold</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
