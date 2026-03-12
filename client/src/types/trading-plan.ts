export interface VolumeLevel {
  price: number;
  volume: number;
  type: 'high' | 'low';
}

export interface Technicals {
  ma20: number;
  ma50: number;
  ma200: number;
  rsi: number;
  volumeProfile: VolumeLevel[];
}

export interface Levels {
  support: number[];
  resistance: number[];
}

export interface Targets {
  pt1: number;
  pt2: number;
  pt3: number;
}

export interface EntryZone {
  low: number;
  high: number;
}

export interface Trade {
  bias: 'bullish' | 'bearish' | 'neutral';
  entryZone: EntryZone;
  stopLoss: number;
  targets: Targets;
  riskRewardRatio: number;
}

export interface Sentiment {
  score: number; // -100 to 100
  institutionalFlow: string;
  smartMoneyBets: string[];
}

export interface TradingPlan {
  ticker: string;
  analyzedAt: string;
  technicals: Technicals;
  levels: Levels;
  trade: Trade;
  sentiment: Sentiment;
  summary: string;
}
