import type { TradingPlan } from './trading-plan';

export interface ScoreBreakdown {
  riskRewardScore: number;
  sentimentScore: number;
  rsiScore: number;
  biasScore: number;
}

export interface ScoredOpportunity {
  tradingPlan: TradingPlan;
  score: number;
  scoreBreakdown: ScoreBreakdown;
}

export interface DailyAnalysisResult {
  date: string;
  analyzedCount: number;
  opportunities: ScoredOpportunity[];
  errors?: string[];
}

export interface Position {
  id: string;
  userId: string;
  ticker: string;
  entryPrice: number;
  quantity: number;
  stopLoss: number;
  takeProfit: number;
  status: 'open' | 'closed';
  etoroId: string;
  openedAt: string;
  closedAt?: string;
  closePrice?: number;
  realizedPnl?: number;
}

export interface Transaction {
  id: string;
  userId: string;
  type: 'BUY' | 'SELL';
  ticker: string;
  price: number;
  quantity: number;
  total: number;
  pnl?: number;
  timestamp: string;
}

export interface TradingRules {
  maxPositionPercent: number;
  maxConcurrentPositions: number;
  minScoreThreshold: number;
  dailyLossLimit: number;
}

export interface Portfolio {
  userId: string;
  budget: number;
  availableBalance: number;
  maxPositionPercent: number;
  maxConcurrentPositions: number;
  minScoreThreshold: number;
  dailyLossLimit: number;
  createdAt: string;
  updatedAt: string;
}

export interface PortfolioSummary {
  budget: number;
  availableBalance: number;
  investedAmount: number;
  openPositions: Position[];
  dailyPnl: number;
  totalPnl: number;
  rules: TradingRules;
}

export interface ExecutionResult {
  success: boolean;
  executedCount: number;
  skippedCount: number;
  errors?: string[];
  positions?: Position[];
}

export const DEFAULT_TRADING_RULES: TradingRules = {
  maxPositionPercent: 0.05,
  maxConcurrentPositions: 5,
  minScoreThreshold: 60,
  dailyLossLimit: 0.10,
};
