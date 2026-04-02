import { useMemo, useState } from 'react';
import { useOpportunities, useLivePrices } from '../../hooks';
import { recalculateScore, sortByAdjustedScore, type AdjustedOpportunity, type PriceStatus } from '../../lib/scorer';

type SortMode = 'ai' | 'live';
type LayoutMode = 'row' | 'card';

function getScoreColor(score: number): string {
  if (score >= 70) return 'bg-green-500';
  if (score >= 40) return 'bg-yellow-500';
  return 'bg-red-500';
}

function getScoreTextColor(score: number): string {
  if (score >= 70) return 'text-green-400';
  if (score >= 40) return 'text-yellow-400';
  return 'text-red-400';
}

function getBiasColor(bias: string): string {
  if (bias === 'bullish') return 'text-green-400';
  if (bias === 'bearish') return 'text-red-400';
  return 'text-gray-400';
}

function getBiasBgColor(bias: string): string {
  if (bias === 'bullish') return 'bg-green-500/20 text-green-400';
  if (bias === 'bearish') return 'bg-red-500/20 text-red-400';
  return 'bg-gray-500/20 text-gray-400';
}

function getPriceStatusInfo(status: PriceStatus): { label: string; color: string } {
  switch (status) {
    case 'in_zone':
      return { label: 'IN ZONE', color: 'bg-green-500' };
    case 'above_zone':
      return { label: 'ABOVE', color: 'bg-blue-500' };
    case 'below_zone':
      return { label: 'BELOW', color: 'bg-blue-500' };
    case 'triggered':
      return { label: 'TARGET HIT', color: 'bg-purple-500' };
    case 'stopped_out':
      return { label: 'STOPPED', color: 'bg-red-600' };
    default:
      return { label: '', color: '' };
  }
}

interface OpportunityCardProps {
  opportunity: AdjustedOpportunity;
  onSelect: (ticker: string) => void;
  rankChange?: number; // positive = moved up, negative = moved down
}

function OpportunityCard({ opportunity, onSelect, rankChange }: OpportunityCardProps) {
  const { tradingPlan, score, scoreBreakdown, adjustedScore, priceStatus, currentPrice, priceChange, liveRR } = opportunity;
  const { trade, technicals } = tradingPlan;
  const statusInfo = getPriceStatusInfo(priceStatus);
  const hasLivePrice = currentPrice !== undefined;
  const displayScore = hasLivePrice ? adjustedScore : score;
  const hasLiveRR = liveRR !== undefined && liveRR > 0;

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
          {statusInfo.label && (
            <span className={`${statusInfo.color} text-white text-xs font-bold px-2 py-0.5 rounded`}>
              {statusInfo.label}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {rankChange !== undefined && rankChange !== 0 && (
            <span className={`text-xs font-medium ${rankChange > 0 ? 'text-green-400' : 'text-red-400'}`}>
              {rankChange > 0 ? `↑${rankChange}` : `↓${Math.abs(rankChange)}`}
            </span>
          )}
          {hasLivePrice && adjustedScore !== score && (
            <span className="text-gray-500 text-xs line-through">{score.toFixed(0)}</span>
          )}
          <div className={`${getScoreColor(displayScore)} text-white text-sm font-bold px-3 py-1 rounded-full`}>
            {displayScore.toFixed(0)}
          </div>
        </div>
      </div>

      {/* Live Price Section */}
      {hasLivePrice && (
        <div className="bg-gray-900 rounded p-2 mb-3">
          <div className="flex justify-between items-center">
            <span className="text-gray-400 text-xs">Live Price</span>
            <div className="flex items-center gap-2">
              <span className="text-white font-mono text-sm">${currentPrice.toFixed(2)}</span>
              <span className={`text-xs ${priceChange >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                {priceChange >= 0 ? '↑' : '↓'} {Math.abs(priceChange).toFixed(2)}%
              </span>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-2 gap-4 text-sm">
        <div>
          <p className="text-gray-400">Entry Zone</p>
          <p className="text-white">${trade.entryZone.low.toFixed(2)} - ${trade.entryZone.high.toFixed(2)}</p>
        </div>
        <div>
          <p className="text-gray-400">R/R Ratio</p>
          <div className="flex items-center gap-2">
            {hasLiveRR && liveRR !== trade.riskRewardRatio && (
              <span className="text-gray-500 text-xs line-through">{trade.riskRewardRatio.toFixed(2)}:1</span>
            )}
            <p className={hasLiveRR ? (liveRR >= 2 ? 'text-green-400' : liveRR >= 1 ? 'text-yellow-400' : 'text-red-400') : 'text-white'}>
              {hasLiveRR ? liveRR.toFixed(2) : trade.riskRewardRatio.toFixed(2)}:1
            </p>
          </div>
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

interface OpportunityRowProps {
  opportunity: AdjustedOpportunity;
  onSelect: (ticker: string) => void;
  rankChange?: number;
}

function OpportunityRow({ opportunity, onSelect, rankChange }: OpportunityRowProps) {
  const { tradingPlan, score, scoreBreakdown, adjustedScore, priceStatus, currentPrice, priceChange, liveRR } = opportunity;
  const { trade, technicals } = tradingPlan;
  const statusInfo = getPriceStatusInfo(priceStatus);
  const hasLivePrice = currentPrice !== undefined;
  const displayScore = hasLivePrice ? adjustedScore : score;
  const hasLiveRR = liveRR !== undefined && liveRR > 0;
  const displayRR = hasLiveRR ? liveRR : trade.riskRewardRatio;
  
  // Show price: currentPrice (live or initial) > entry zone midpoint as fallback
  const displayPrice = currentPrice ?? ((trade.entryZone.low + trade.entryZone.high) / 2);

  return (
    <div 
      className="bg-gray-800 hover:bg-gray-750 cursor-pointer transition-colors border-b border-gray-700 last:border-b-0"
      onClick={() => onSelect(tradingPlan.ticker)}
    >
      {/* Mobile Layout (stacked) */}
      <div className="block sm:hidden p-3">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <span className="text-white font-bold">{tradingPlan.ticker}</span>
            <span className={`text-xs px-1.5 py-0.5 rounded ${getBiasBgColor(trade.bias)}`}>
              {trade.bias.charAt(0).toUpperCase()}
            </span>
            {statusInfo.label && (
              <span className={`${statusInfo.color} text-white text-xs font-medium px-1.5 py-0.5 rounded`}>
                {statusInfo.label}
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            {rankChange !== undefined && rankChange !== 0 && (
              <span className={`text-xs ${rankChange > 0 ? 'text-green-400' : 'text-red-400'}`}>
                {rankChange > 0 ? `↑${rankChange}` : `↓${Math.abs(rankChange)}`}
              </span>
            )}
            {hasLivePrice && adjustedScore !== score && (
              <span className="text-gray-500 text-xs line-through">{score.toFixed(0)}</span>
            )}
            <span className={`font-bold ${getScoreTextColor(displayScore)}`}>
              {displayScore.toFixed(0)}
            </span>
          </div>
        </div>
        
        {/* Price row */}
        <div className="flex items-center justify-between text-xs mb-2">
          <span className="text-gray-400">Ask Price</span>
          <div className="flex items-center gap-2">
            <span className={`font-mono ${hasLivePrice ? 'text-white' : 'text-gray-500'}`}>
              ${displayPrice.toFixed(2)}
            </span>
            {hasLivePrice && priceChange !== undefined && (
              <span className={`${priceChange >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                {priceChange >= 0 ? '↑' : '↓'}{Math.abs(priceChange).toFixed(1)}%
              </span>
            )}
          </div>
        </div>

        {/* Trade info row */}
        <div className="grid grid-cols-2 gap-2 text-xs mb-2">
          <div>
            <span className="text-gray-500">Entry: </span>
            <span className="text-gray-300">${trade.entryZone.low.toFixed(2)} - ${trade.entryZone.high.toFixed(2)}</span>
          </div>
          <div className="text-right">
            <span className="text-gray-500">R/R: </span>
            {hasLiveRR && liveRR !== trade.riskRewardRatio && (
              <span className="text-gray-600 line-through mr-1">{trade.riskRewardRatio.toFixed(1)}:1</span>
            )}
            <span className={hasLiveRR ? (displayRR >= 2 ? 'text-green-400' : displayRR >= 1 ? 'text-yellow-400' : 'text-red-400') : 'text-gray-300'}>
              {displayRR.toFixed(2)}:1
            </span>
          </div>
          <div>
            <span className="text-gray-500">Stop: </span>
            <span className="text-red-400">${trade.stopLoss.toFixed(2)}</span>
          </div>
          <div className="text-right">
            <span className="text-gray-500">Target: </span>
            <span className="text-green-400">${trade.targets.pt1.toFixed(2)}</span>
          </div>
        </div>

        {/* Technical indicators */}
        <div className="flex justify-between text-xs text-gray-500 pt-2 border-t border-gray-700">
          <span>RSI: {technicals.rsi.toFixed(0)}</span>
          <span>R/R: {scoreBreakdown.riskRewardScore.toFixed(0)}</span>
          <span>Sent: {scoreBreakdown.sentimentScore.toFixed(0)}</span>
        </div>
      </div>

      {/* Desktop Layout (horizontal) */}
      <div className="hidden sm:flex items-center px-4 py-3 gap-3">
        {/* Ticker & Bias */}
        <div className="w-24 flex items-center gap-2">
          <span className="text-white font-bold">{tradingPlan.ticker}</span>
          <span className={`text-xs px-1.5 py-0.5 rounded ${getBiasBgColor(trade.bias)}`}>
            {trade.bias.charAt(0).toUpperCase()}
          </span>
        </div>

        {/* Score */}
        <div className="w-20 flex items-center gap-1">
          {rankChange !== undefined && rankChange !== 0 && (
            <span className={`text-xs ${rankChange > 0 ? 'text-green-400' : 'text-red-400'}`}>
              {rankChange > 0 ? '↑' : '↓'}
            </span>
          )}
          {hasLivePrice && adjustedScore !== score && (
            <span className="text-gray-500 text-xs line-through">{score.toFixed(0)}</span>
          )}
          <span className={`font-bold ${getScoreTextColor(displayScore)}`}>
            {displayScore.toFixed(0)}
          </span>
        </div>

        {/* Status */}
        <div className="w-20">
          {statusInfo.label && (
            <span className={`${statusInfo.color} text-white text-xs font-medium px-2 py-0.5 rounded`}>
              {statusInfo.label}
            </span>
          )}
        </div>

        {/* Ask Price */}
        <div className="w-28 text-right flex items-center justify-end gap-1">
          <span className={`font-mono text-sm ${hasLivePrice ? 'text-white' : 'text-gray-500'}`}>
            ${displayPrice.toFixed(2)}
          </span>
          {hasLivePrice && priceChange !== undefined && (
            <span className={`text-xs ${priceChange >= 0 ? 'text-green-400' : 'text-red-400'}`}>
              {priceChange >= 0 ? '↑' : '↓'}{Math.abs(priceChange).toFixed(1)}%
            </span>
          )}
        </div>

        {/* Entry Zone */}
        <div className="w-36 text-gray-400 text-sm">
          ${trade.entryZone.low.toFixed(2)} - ${trade.entryZone.high.toFixed(2)}
        </div>

        {/* Stop Loss */}
        <div className="w-20 text-red-400 text-sm">
          ${trade.stopLoss.toFixed(2)}
        </div>

        {/* Target */}
        <div className="w-20 text-green-400 text-sm">
          ${trade.targets.pt1.toFixed(2)}
        </div>

        {/* R/R */}
        <div className="w-20 flex items-center gap-1">
          {hasLiveRR && liveRR !== trade.riskRewardRatio && (
            <span className="text-gray-500 text-xs line-through">{trade.riskRewardRatio.toFixed(1)}</span>
          )}
          <span className={`text-sm ${hasLiveRR ? (displayRR >= 2 ? 'text-green-400' : displayRR >= 1 ? 'text-yellow-400' : 'text-red-400') : 'text-white'}`}>
            {displayRR.toFixed(2)}:1
          </span>
        </div>

        {/* RSI */}
        <div className="w-12 text-gray-400 text-sm text-center">
          {technicals.rsi.toFixed(0)}
        </div>

        {/* Score breakdown */}
        <div className="w-24 text-gray-500 text-xs">
          <span className="mr-2">R:{scoreBreakdown.riskRewardScore.toFixed(0)}</span>
          <span>S:{scoreBreakdown.sentimentScore.toFixed(0)}</span>
        </div>
      </div>
    </div>
  );
}

// Table header for row layout
function OpportunityRowHeader() {
  return (
    <div className="hidden sm:flex items-center px-4 py-2 gap-3 bg-gray-900 text-xs text-gray-500 font-medium uppercase tracking-wider border-b border-gray-700">
      <div className="w-24">Ticker</div>
      <div className="w-20">Score</div>
      <div className="w-20">Status</div>
      <div className="w-28 text-right">Ask</div>
      <div className="w-36">Entry Zone</div>
      <div className="w-20">Stop</div>
      <div className="w-20">Target</div>
      <div className="w-20">R/R</div>
      <div className="w-12 text-center">RSI</div>
      <div className="w-24">Scores</div>
    </div>
  );
}

interface OpportunitiesViewProps {
  onSelectTicker?: (ticker: string) => void;
}

export function OpportunitiesView({ onSelectTicker }: OpportunitiesViewProps) {
  const [sortMode, setSortMode] = useState<SortMode>('ai');
  const [layoutMode, setLayoutMode] = useState<LayoutMode>('row');
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

  // Extract symbols from opportunities for WebSocket subscription
  const symbols = useMemo(() => 
    opportunities.map(opp => opp.tradingPlan.ticker),
    [opportunities]
  );

  // Subscribe to live prices for all opportunity symbols
  const { prices, isConnected, error: wsError } = useLivePrices({ 
    symbols,
    autoConnect: symbols.length > 0
  });

  // Recalculate scores with live prices (keep AI order)
  const adjustedOpportunities = useMemo(() => {
    return opportunities.map(opp => 
      recalculateScore(opp, prices[opp.tradingPlan.ticker])
    );
  }, [opportunities, prices]);

  // Create live-sorted version for comparison / toggle
  const liveSortedOpportunities = useMemo(() => {
    return sortByAdjustedScore([...adjustedOpportunities]);
  }, [adjustedOpportunities]);

  // Calculate rank changes (AI rank vs live rank)
  const rankChanges = useMemo(() => {
    const changes: Record<string, number> = {};
    adjustedOpportunities.forEach((opp, aiRank) => {
      const liveRank = liveSortedOpportunities.findIndex(
        o => o.tradingPlan.ticker === opp.tradingPlan.ticker
      );
      changes[opp.tradingPlan.ticker] = aiRank - liveRank; // positive = moved up in live
    });
    return changes;
  }, [adjustedOpportunities, liveSortedOpportunities]);

  // Pick which list to display based on sort mode
  const displayOpportunities = sortMode === 'ai' ? adjustedOpportunities : liveSortedOpportunities;

  // Count opportunities by status
  const statusCounts = useMemo(() => {
    const counts = { inZone: 0, belowZone: 0, other: 0 };
    displayOpportunities.forEach(opp => {
      if (opp.priceStatus === 'in_zone') counts.inZone++;
      else if (opp.priceStatus === 'below_zone') counts.belowZone++;
      else counts.other++;
    });
    return counts;
  }, [displayOpportunities]);

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
    <div className="p-4 md:p-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-xl md:text-2xl font-bold text-white">Opportunities</h1>
            {/* Sort mode toggle */}
            {isConnected && displayOpportunities.length > 0 && (
              <div className="flex bg-gray-700 rounded-lg p-0.5">
                <button
                  onClick={() => setSortMode('ai')}
                  className={`px-2 py-1 text-xs font-medium rounded-md transition-colors ${
                    sortMode === 'ai' 
                      ? 'bg-blue-600 text-white' 
                      : 'text-gray-400 hover:text-white'
                  }`}
                >
                  AI Order
                </button>
                <button
                  onClick={() => setSortMode('live')}
                  className={`px-2 py-1 text-xs font-medium rounded-md transition-colors ${
                    sortMode === 'live' 
                      ? 'bg-blue-600 text-white' 
                      : 'text-gray-400 hover:text-white'
                  }`}
                >
                  Live Sort
                </button>
              </div>
            )}
            {/* WebSocket connection status */}
            <div className="flex items-center gap-1.5">
              <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-gray-500'}`} />
              <span className="text-xs text-gray-500">
                {isConnected ? 'Live' : 'Offline'}
              </span>
            </div>
          </div>
          {lastAnalysisDate && (
            <p className="text-gray-400 text-xs md:text-sm">
              {opportunities.length} opportunities from {lastAnalysisDate}
              {isStale && <span className="text-yellow-400 ml-2">(stale)</span>}
            </p>
          )}
          {/* Status counts */}
          {hasExistingData && isConnected && (
            <div className="flex gap-3 mt-1 text-xs">
              {statusCounts.inZone > 0 && (
                <span className="text-green-400">{statusCounts.inZone} in zone</span>
              )}
              {statusCounts.belowZone > 0 && (
                <span className="text-blue-400">{statusCounts.belowZone} below</span>
              )}
            </div>
          )}
          {analysisResult && (
            <p className="text-gray-500 text-xs mt-1">
              {analysisResult.analyzedCount} tickers analyzed
            </p>
          )}
        </div>
        <div className="flex items-center gap-2">
          {/* Layout toggle */}
          {displayOpportunities.length > 0 && (
            <div className="flex bg-gray-700 rounded-lg p-0.5">
              <button
                onClick={() => setLayoutMode('row')}
                className={`p-1.5 rounded-md transition-colors ${
                  layoutMode === 'row' 
                    ? 'bg-gray-600 text-white' 
                    : 'text-gray-400 hover:text-white'
                }`}
                title="Row layout"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                </svg>
              </button>
              <button
                onClick={() => setLayoutMode('card')}
                className={`p-1.5 rounded-md transition-colors ${
                  layoutMode === 'card' 
                    ? 'bg-gray-600 text-white' 
                    : 'text-gray-400 hover:text-white'
                }`}
                title="Card layout"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                </svg>
              </button>
            </div>
          )}
          <button
            onClick={() => runDailyAnalysis()}
            disabled={loading}
            className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 text-white px-3 py-1.5 md:px-4 md:py-2 rounded-lg text-xs md:text-sm font-medium transition-colors"
          >
            {loading ? 'Analyzing...' : hasExistingData ? 'Refresh' : 'Run Analysis'}
          </button>
          {opportunities.length > 0 && (
            <button
              onClick={() => executeTrades()}
              disabled={executing}
              className="bg-green-600 hover:bg-green-700 disabled:bg-gray-600 text-white px-3 py-1.5 md:px-4 md:py-2 rounded-lg text-xs md:text-sm font-medium transition-colors"
            >
              {executing ? '...' : 'Execute'}
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="bg-red-900/50 border border-red-500 text-red-200 px-4 py-3 rounded-lg mb-6">
          {error}
        </div>
      )}

      {wsError && (
        <div className="bg-yellow-900/50 border border-yellow-500 text-yellow-200 px-3 py-2 rounded-lg mb-4 text-sm">
          Live prices unavailable: {wsError}
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

      {displayOpportunities.length === 0 && !loading ? (
        <div className="text-center py-12">
          <p className="text-gray-400 text-lg">No opportunities available</p>
          <p className="text-gray-500 text-sm mt-2">
            Click "Run Daily Analysis" to analyze your watchlist
          </p>
        </div>
      ) : layoutMode === 'row' ? (
        /* Row Layout */
        <div className="bg-gray-800 rounded-lg overflow-hidden">
          <OpportunityRowHeader />
          {displayOpportunities.map((opp) => (
            <OpportunityRow
              key={opp.tradingPlan.ticker}
              opportunity={opp}
              onSelect={handleSelect}
              rankChange={sortMode === 'ai' ? rankChanges[opp.tradingPlan.ticker] : undefined}
            />
          ))}
        </div>
      ) : (
        /* Card Layout */
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {displayOpportunities.map((opp) => (
            <OpportunityCard
              key={opp.tradingPlan.ticker}
              opportunity={opp}
              onSelect={handleSelect}
              rankChange={sortMode === 'ai' ? rankChanges[opp.tradingPlan.ticker] : undefined}
            />
          ))}
        </div>
      )}
    </div>
  );
}
