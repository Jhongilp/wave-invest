import type { TradingPlan as TradingPlanType } from '../../types';

interface TradingPlanViewProps {
  plan: TradingPlanType | null;
  loading: boolean;
  error: Error | null;
}

export function TradingPlanView({ plan, loading, error }: TradingPlanViewProps) {
  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto" />
          <p className="mt-4 text-gray-400">Analyzing with AI...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center text-red-400">
          <p>Error: {error.message}</p>
        </div>
      </div>
    );
  }

  if (!plan) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center text-gray-500">
          <p className="text-xl">Select a ticker to analyze</p>
          <p className="text-sm mt-2">Click on any ticker in the watchlist to generate a trading plan</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-auto p-6">
      <div className="max-w-4xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">{plan.ticker}</h1>
          <span
            className={`px-4 py-1 rounded-full text-sm font-medium ${
              plan.trade.bias === 'bullish'
                ? 'bg-green-500/20 text-green-400'
                : plan.trade.bias === 'bearish'
                ? 'bg-red-500/20 text-red-400'
                : 'bg-gray-500/20 text-gray-400'
            }`}
          >
            {plan.trade.bias.toUpperCase()}
          </span>
        </div>

        {/* Summary */}
        <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-700">
          <h3 className="text-sm font-medium text-gray-400 mb-2">AI Summary</h3>
          <p className="text-white">{plan.summary}</p>
        </div>

        {/* Technical Indicators */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-700">
            <h3 className="text-sm font-medium text-gray-400 mb-3">Moving Averages</h3>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-gray-300">MA 20</span>
                <span className="text-white font-medium">${plan.technicals.ma20.toFixed(2)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-300">MA 50</span>
                <span className="text-white font-medium">${plan.technicals.ma50.toFixed(2)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-300">MA 200</span>
                <span className="text-white font-medium">${plan.technicals.ma200.toFixed(2)}</span>
              </div>
            </div>
          </div>

          <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-700">
            <h3 className="text-sm font-medium text-gray-400 mb-3">RSI</h3>
            <div className="flex items-center gap-4">
              <div className="flex-1 bg-gray-700 rounded-full h-3">
                <div
                  className={`h-3 rounded-full ${
                    plan.technicals.rsi > 70
                      ? 'bg-red-500'
                      : plan.technicals.rsi < 30
                      ? 'bg-green-500'
                      : 'bg-blue-500'
                  }`}
                  style={{ width: `${plan.technicals.rsi}%` }}
                />
              </div>
              <span className="text-white font-medium w-12">{plan.technicals.rsi.toFixed(1)}</span>
            </div>
            <div className="flex justify-between text-xs text-gray-500 mt-1">
              <span>Oversold</span>
              <span>Neutral</span>
              <span>Overbought</span>
            </div>
          </div>
        </div>

        {/* Support/Resistance Levels */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-700">
            <h3 className="text-sm font-medium text-green-400 mb-3">Support Levels</h3>
            <div className="space-y-2">
              {plan.levels.support.map((level, i) => (
                <div key={i} className="flex justify-between">
                  <span className="text-gray-300">S{i + 1}</span>
                  <span className="text-green-400 font-medium">${level.toFixed(2)}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-700">
            <h3 className="text-sm font-medium text-red-400 mb-3">Resistance Levels</h3>
            <div className="space-y-2">
              {plan.levels.resistance.map((level, i) => (
                <div key={i} className="flex justify-between">
                  <span className="text-gray-300">R{i + 1}</span>
                  <span className="text-red-400 font-medium">${level.toFixed(2)}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Trade Setup */}
        <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-700">
          <h3 className="text-sm font-medium text-gray-400 mb-4">Trade Setup</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <span className="text-xs text-gray-500">Entry Zone</span>
              <p className="text-white font-medium">
                ${plan.trade.entryZone.low.toFixed(2)} - ${plan.trade.entryZone.high.toFixed(2)}
              </p>
            </div>
            <div>
              <span className="text-xs text-gray-500">Stop Loss</span>
              <p className="text-red-400 font-medium">${plan.trade.stopLoss.toFixed(2)}</p>
            </div>
            <div>
              <span className="text-xs text-gray-500">Risk/Reward</span>
              <p className="text-blue-400 font-medium">{plan.trade.riskRewardRatio.toFixed(1)}:1</p>
            </div>
            <div>
              <span className="text-xs text-gray-500">Bias</span>
              <p className={`font-medium ${
                plan.trade.bias === 'bullish' ? 'text-green-400' : plan.trade.bias === 'bearish' ? 'text-red-400' : 'text-gray-400'
              }`}>
                {plan.trade.bias.charAt(0).toUpperCase() + plan.trade.bias.slice(1)}
              </p>
            </div>
          </div>
        </div>

        {/* Price Targets */}
        <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-700">
          <h3 className="text-sm font-medium text-gray-400 mb-4">Price Targets</h3>
          <div className="grid grid-cols-3 gap-4">
            <div className="text-center p-3 bg-gray-700/50 rounded-lg">
              <span className="text-xs text-gray-500">PT1</span>
              <p className="text-green-400 font-semibold text-lg">${plan.trade.targets.pt1.toFixed(2)}</p>
            </div>
            <div className="text-center p-3 bg-gray-700/50 rounded-lg">
              <span className="text-xs text-gray-500">PT2</span>
              <p className="text-green-400 font-semibold text-lg">${plan.trade.targets.pt2.toFixed(2)}</p>
            </div>
            <div className="text-center p-3 bg-gray-700/50 rounded-lg">
              <span className="text-xs text-gray-500">PT3</span>
              <p className="text-green-400 font-semibold text-lg">${plan.trade.targets.pt3.toFixed(2)}</p>
            </div>
          </div>
        </div>

        {/* Sentiment */}
        <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-700">
          <h3 className="text-sm font-medium text-gray-400 mb-4">Sentiment Analysis</h3>
          <div className="flex items-center gap-4 mb-4">
            <div className="flex-1">
              <div className="flex justify-between text-xs text-gray-500 mb-1">
                <span>Bearish</span>
                <span>Neutral</span>
                <span>Bullish</span>
              </div>
              <div className="bg-gray-700 rounded-full h-2 relative">
                <div
                  className="absolute top-1/2 -translate-y-1/2 w-4 h-4 rounded-full bg-blue-500 border-2 border-white"
                  style={{ left: `calc(${(plan.sentiment.score + 100) / 2}% - 8px)` }}
                />
              </div>
            </div>
            <span className="text-white font-medium">{plan.sentiment.score}</span>
          </div>
          <p className="text-gray-300 text-sm mb-3">{plan.sentiment.institutionalFlow}</p>
          <div className="space-y-1">
            {plan.sentiment.smartMoneyBets.map((bet, i) => (
              <div key={i} className="flex items-center gap-2 text-sm text-gray-400">
                <span className="text-blue-400">•</span>
                {bet}
              </div>
            ))}
          </div>
        </div>

        {/* Timestamp */}
        <p className="text-xs text-gray-500 text-center">
          Analyzed at {new Date(plan.analyzedAt).toLocaleString()}
        </p>
      </div>
    </div>
  );
}
