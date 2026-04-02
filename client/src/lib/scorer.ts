import type { ScoredOpportunity, ScoreBreakdown } from '../types/portfolio';
import type { TradingPlan, Trade, Technicals } from '../types/trading-plan';
import type { LivePrice } from '../types/ticker';

// Scoring weights (matching backend)
const WEIGHT_RISK_REWARD = 0.40;
const WEIGHT_SENTIMENT = 0.25;
const WEIGHT_RSI = 0.20;
const WEIGHT_BIAS = 0.15;

// Price adjustment weight (how much live price affects score)
const WEIGHT_PRICE_PROXIMITY = 0.30; // 30% adjustment based on live price

/**
 * Recalculates opportunity score based on live price data.
 * The base score from AI analysis is adjusted based on how close
 * the current price is to the entry zone.
 */
export function recalculateScore(
  opportunity: ScoredOpportunity,
  livePrice: LivePrice | undefined
): AdjustedOpportunity {
  if (!livePrice) {
    return {
      ...opportunity,
      adjustedScore: opportunity.score,
      priceStatus: 'unknown',
      currentPrice: undefined,
      priceChange: 0,
      liveRR: undefined
    };
  }

  const { tradingPlan } = opportunity;
  const trade = tradingPlan.trade;
  
  // Use ask price (buy price) - fallback to last if ask not available
  let currentPrice: number;
  if (livePrice.ask > 0) {
    currentPrice = livePrice.ask;
  } else if (livePrice.last > 0) {
    currentPrice = livePrice.last;
  } else {
    // No valid price data
    return {
      ...opportunity,
      adjustedScore: opportunity.score,
      priceStatus: 'unknown',
      currentPrice: undefined,
      priceChange: 0,
      liveRR: undefined
    };
  }
  
  // Determine price status relative to entry zone
  const priceStatus = getPriceStatus(currentPrice, trade);
  
  // Calculate price proximity score (0-100)
  const proximityScore = calculateProximityScore(currentPrice, trade);
  
  // Adjust the original score based on price proximity
  // If price is in entry zone, boost score; if far away, reduce it
  const baseScore = opportunity.score;
  const adjustedScore = baseScore * (1 - WEIGHT_PRICE_PROXIMITY) + 
                       proximityScore * WEIGHT_PRICE_PROXIMITY;
  
  // Calculate price change from entry zone midpoint
  const entryMid = (trade.entryZone.low + trade.entryZone.high) / 2;
  const priceChange = ((currentPrice - entryMid) / entryMid) * 100;

  // Calculate live R/R based on current ask price
  const liveRR = calculateLiveRR(currentPrice, trade);

  return {
    ...opportunity,
    adjustedScore: Math.min(100, Math.max(0, adjustedScore)),
    priceStatus,
    currentPrice,
    priceChange,
    liveRR
  };
}

/**
 * Determines the status of current price relative to entry zone
 */
export type PriceStatus = 
  | 'in_zone'      // Price is within entry zone
  | 'above_zone'   // Price is above entry zone
  | 'below_zone'   // Price is below entry zone
  | 'triggered'    // Price hit target
  | 'stopped_out'  // Price hit stop loss
  | 'unknown';

function getPriceStatus(currentPrice: number, trade: Trade): PriceStatus {
  const { entryZone, stopLoss, targets, bias } = trade;
  
  // Check if stopped out (with sanity checks for consistent data)
  if (bias === 'bullish' && stopLoss < entryZone.low && currentPrice <= stopLoss) {
    return 'stopped_out';
  }
  if (bias === 'bearish' && stopLoss > entryZone.high && currentPrice >= stopLoss) {
    return 'stopped_out';
  }
  
  // Check if target hit (PT1) - with sanity checks
  if (bias === 'bullish' && targets.pt1 > entryZone.high && currentPrice >= targets.pt1) {
    return 'triggered';
  }
  if (bias === 'bearish' && targets.pt1 < entryZone.low && currentPrice <= targets.pt1) {
    return 'triggered';
  }
  
  // Check entry zone
  if (currentPrice >= entryZone.low && currentPrice <= entryZone.high) {
    return 'in_zone';
  }
  
  // Price is outside entry zone
  if (currentPrice < entryZone.low) {
    return 'below_zone';
  }
  
  if (currentPrice > entryZone.high) {
    return 'above_zone';
  }
  
  return 'unknown';
}

/**
 * Calculates live Risk/Reward ratio based on current price
 */
function calculateLiveRR(currentPrice: number, trade: Trade): number | undefined {
  const { stopLoss, targets, bias } = trade;
  
  let risk: number;
  let reward: number;
  
  if (bias === 'bullish') {
    // Bullish: risk is distance to stop (below), reward is distance to target (above)
    risk = currentPrice - stopLoss;
    reward = targets.pt1 - currentPrice;
  } else {
    // Bearish: risk is distance to stop (above), reward is distance to target (below)
    risk = stopLoss - currentPrice;
    reward = currentPrice - targets.pt1;
  }
  
  // Invalid if risk or reward is negative or zero
  if (risk <= 0 || reward <= 0) {
    return undefined;
  }
  
  return reward / risk;
}

/**
 * Calculates a score (0-100) based on how close the price is to the entry zone
 */
function calculateProximityScore(currentPrice: number, trade: Trade): number {
  const { entryZone, stopLoss, targets, bias } = trade;
  
  // If in entry zone, full score
  if (currentPrice >= entryZone.low && currentPrice <= entryZone.high) {
    return 100;
  }
  
  if (bias === 'bullish') {
    // For bullish: below entry is good (can still enter), above is less good
    if (currentPrice < entryZone.low) {
      // Below entry zone - still good if above stop loss
      const distanceFromStop = currentPrice - stopLoss;
      const rangeToStop = entryZone.low - stopLoss;
      if (rangeToStop > 0) {
        const ratio = distanceFromStop / rangeToStop;
        return Math.max(0, ratio * 100);
      }
    } else {
      // Above entry zone - score decreases as we approach PT1
      const distanceFromEntry = currentPrice - entryZone.high;
      const rangeToTarget = targets.pt1 - entryZone.high;
      if (rangeToTarget > 0) {
        const ratio = 1 - (distanceFromEntry / rangeToTarget);
        return Math.max(0, ratio * 80); // Cap at 80 when above entry
      }
    }
  } else if (bias === 'bearish') {
    // For bearish: above entry is good (can still enter), below is less good
    if (currentPrice > entryZone.high) {
      // Above entry zone - still good if below stop loss
      const distanceFromStop = stopLoss - currentPrice;
      const rangeToStop = stopLoss - entryZone.high;
      if (rangeToStop > 0) {
        const ratio = distanceFromStop / rangeToStop;
        return Math.max(0, ratio * 100);
      }
    } else {
      // Below entry zone - score decreases as we approach PT1
      const distanceFromEntry = entryZone.low - currentPrice;
      const rangeToTarget = entryZone.low - targets.pt1;
      if (rangeToTarget > 0) {
        const ratio = 1 - (distanceFromEntry / rangeToTarget);
        return Math.max(0, ratio * 80); // Cap at 80 when below entry
      }
    }
  }
  
  return 50; // Default for neutral or edge cases
}

/**
 * Extended opportunity with live price adjustments
 */
export interface AdjustedOpportunity extends ScoredOpportunity {
  adjustedScore: number;
  priceStatus: PriceStatus;
  currentPrice: number | undefined;
  priceChange: number; // Percentage change from entry zone mid
  liveRR: number | undefined; // Live risk/reward based on current price
}

/**
 * Sorts opportunities by adjusted score (descending)
 */
export function sortByAdjustedScore(opportunities: AdjustedOpportunity[]): AdjustedOpportunity[] {
  return [...opportunities].sort((a, b) => b.adjustedScore - a.adjustedScore);
}

/**
 * Filters opportunities based on price status
 */
export function filterByPriceStatus(
  opportunities: AdjustedOpportunity[],
  statuses: PriceStatus[]
): AdjustedOpportunity[] {
  return opportunities.filter(opp => statuses.includes(opp.priceStatus));
}

/**
 * Gets actionable opportunities (in zone)
 */
export function getActionableOpportunities(opportunities: AdjustedOpportunity[]): AdjustedOpportunity[] {
  return filterByPriceStatus(opportunities, ['in_zone']);
}

/**
 * Recalculates the base score breakdown (matches backend scorer)
 * This is the same algorithm as Go's scorer.go
 */
export function calculateScoreBreakdown(plan: TradingPlan): ScoreBreakdown {
  return {
    riskRewardScore: scoreRiskReward(plan.trade),
    sentimentScore: scoreSentiment(plan),
    rsiScore: scoreRSI(plan.technicals, plan.trade.bias),
    biasScore: scoreBias(plan.trade.bias)
  };
}

function scoreRiskReward(trade: Trade): number {
  const rr = trade.riskRewardRatio;
  if (rr <= 0) return 0;
  // Scale: 1:1 = 20, 2:1 = 40, 3:1 = 60, 4:1 = 80, 5:1+ = 100
  return Math.min(rr * 20, 100);
}

function scoreSentiment(plan: TradingPlan): number {
  let sentiment = plan.sentiment.score;
  // For bullish bias, positive sentiment is good
  // For bearish bias, negative sentiment is good
  if (plan.trade.bias === 'bearish') {
    sentiment = -sentiment;
  }
  // Convert -100..100 to 0..100
  return (sentiment + 100) / 2;
}

function scoreRSI(technicals: Technicals, bias: string): number {
  const rsi = technicals.rsi;

  if (bias === 'bullish') {
    // For longs, prefer oversold conditions (RSI < 30 is ideal)
    if (rsi <= 30) return 100;
    if (rsi <= 50) return 100 - ((rsi - 30) * 2.5);
    if (rsi <= 70) return 50 - ((rsi - 50) * 2.5);
    return 0;
  } else if (bias === 'bearish') {
    // For shorts, prefer overbought conditions (RSI > 70 is ideal)
    if (rsi >= 70) return 100;
    if (rsi >= 50) return (rsi - 50) * 5;
    return 0;
  }

  // Neutral bias - middle RSI is okay
  return 50;
}

function scoreBias(bias: string): number {
  switch (bias) {
    case 'bullish':
    case 'bearish':
      return 100;
    default:
      return 0;
  }
}

/**
 * Calculates total score from breakdown (matches backend)
 */
export function calculateTotalScore(breakdown: ScoreBreakdown): number {
  return (
    breakdown.riskRewardScore * WEIGHT_RISK_REWARD +
    breakdown.sentimentScore * WEIGHT_SENTIMENT +
    breakdown.rsiScore * WEIGHT_RSI +
    breakdown.biasScore * WEIGHT_BIAS
  );
}
