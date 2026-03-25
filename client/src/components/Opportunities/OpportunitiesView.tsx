import { useOpportunities } from '../../hooks';
import type { ScoredOpportunity } from '../../types';

function getScoreColor(score: number): string {
  if (score >= 70) return 'bg-green-500';
  if (score >= 40) return 'bg-yellow-500';
  return 'bg-red-500';
}

function getBiasColor(bias: string): string {
  if (bias === 'bullish') return 'text-green-400';
  if (bias === 'bearish') return 'text-red-400';
  return 'text-gray-400';
}

interface OpportunityCardProps {
  opportunity: ScoredOpportunity;
  onSelect: (ticker: string) => void;
}

function OpportunityCard({ opportunity, onSelect }: OpportunityCardProps) {
  const { tradingPlan, score, scoreBreakdown } = opportunity;
  const { trade, technicals } = tradingPlan;

  return (
    <div 
      className="bg-gray-800 rounded-lg p-4 hover:bg-gray-750 cursor-pointer transition-colors"
      onClick={() => onSelect(tradingPlan.ticker)}
    >
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-3">
          <h3 className="text-lg font-bold text-white">{tradingPlan.ticker}</h3>
          <span className={`text-sm font-medium ${getBiasColor(trade.bias)}`}>
            {trade.bias.toUpperCase()}
          </span>
        </div>
        <div className={`${getScoreColor(score)} text-white text-sm font-bold px-3 py-1 rounded-full`}>
          {score.toFixed(0)}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 text-sm">
        <div>
          <p className="text-gray-400">Entry Zone</p>
          <p className="text-white">${trade.entryZone.low.toFixed(2)} - ${trade.entryZone.high.toFixed(2)}</p>
        </div>
        <div>
          <p className="text-gray-400">R/R Ratio</p>
          <p className="text-white">{trade.riskRewardRatio.toFixed(2)}:1</p>
        </div>
        <div>
          <p className="text-gray-400">Stop Loss</p>
          <p className="text-red-400">${trade.stopLoss.toFixed(2)}</p>
        </div>
        <div>
          <p className="text-gray-400">Target (PT1)</p>
          <p className="text-green-400">${trade.targets.pt1.toFixed(2)}</p>
        </div>
      </div>

      <div className="mt-3 pt-3 border-t border-gray-700">
        <div className="flex justify-between text-xs text-gray-500">
          <span>RSI: {technicals.rsi.toFixed(0)}</span>
          <span>R/R: {scoreBreakdown.riskRewardScore.toFixed(0)}</span>
          <span>Sent: {scoreBreakdown.sentimentScore.toFixed(0)}</span>
        </div>
      </div>
    </div>
  );
}

interface OpportunitiesViewProps {
  onSelectTicker?: (ticker: string) => void;
}

export function OpportunitiesView({ onSelectTicker }: OpportunitiesViewProps) {
  const {
    opportunities,
    analysisResult,
    executionResult,
    loading,
    initialLoading,
    executing,
    error,
    lastAnalysisDate,
    runDailyAnalysis,
    executeTrades,
  } = useOpportunities();

  const handleSelect = (ticker: string) => {
    if (onSelectTicker) {
      onSelectTicker(ticker);
    }
  };

  const hasExistingData = opportunities.length > 0;
  const today = new Date().toISOString().split('T')[0];
  const isStale = lastAnalysisDate && lastAnalysisDate !== today;

  if (initialLoading) {
    return (
      <div className="p-6">
        <div className="text-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-400">Loading opportunities...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white">Opportunities</h1>
          {lastAnalysisDate && (
            <p className="text-gray-400 text-sm">
              {opportunities.length} opportunities from {lastAnalysisDate}
              {isStale && <span className="text-yellow-400 ml-2">(stale - not from today)</span>}
            </p>
          )}
          {analysisResult && (
            <p className="text-gray-500 text-xs mt-1">
              {analysisResult.analyzedCount} tickers analyzed
            </p>
          )}
        </div>
        <div className="flex gap-3">
          <button
            onClick={() => runDailyAnalysis()}
            disabled={loading}
            className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 text-white px-4 py-2 rounded-lg font-medium transition-colors"
          >
            {loading ? 'Analyzing...' : hasExistingData ? 'Refresh Analysis' : 'Run Daily Analysis'}
          </button>
          {opportunities.length > 0 && (
            <button
              onClick={() => executeTrades()}
              disabled={executing}
              className="bg-green-600 hover:bg-green-700 disabled:bg-gray-600 text-white px-4 py-2 rounded-lg font-medium transition-colors"
            >
              {executing ? 'Executing...' : 'Execute Trades'}
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="bg-red-900/50 border border-red-500 text-red-200 px-4 py-3 rounded-lg mb-6">
          {error}
        </div>
      )}

      {executionResult && (
        <div className="bg-green-900/50 border border-green-500 text-green-200 px-4 py-3 rounded-lg mb-6">
          Executed {executionResult.executedCount} trades, skipped {executionResult.skippedCount}
          {executionResult.errors && executionResult.errors.length > 0 && (
            <ul className="mt-2 text-sm">
              {executionResult.errors.map((err, i) => (
                <li key={i}>• {err}</li>
              ))}
            </ul>
          )}
        </div>
      )}

      {analysisResult?.errors && analysisResult.errors.length > 0 && (
        <div className="bg-yellow-900/50 border border-yellow-500 text-yellow-200 px-4 py-3 rounded-lg mb-6">
          <p className="font-medium">Analysis completed with warnings:</p>
          <ul className="mt-2 text-sm">
            {analysisResult.errors.map((err, i) => (
              <li key={i}>• {err}</li>
            ))}
          </ul>
        </div>
      )}

      {opportunities.length === 0 && !loading ? (
        <div className="text-center py-12">
          <p className="text-gray-400 text-lg">No opportunities available</p>
          <p className="text-gray-500 text-sm mt-2">
            Click "Run Daily Analysis" to analyze your watchlist
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {opportunities.map((opp) => (
            <OpportunityCard
              key={opp.tradingPlan.ticker}
              opportunity={opp}
              onSelect={handleSelect}
            />
          ))}
        </div>
      )}
    </div>
  );
}
