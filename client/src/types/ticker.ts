export interface Ticker {
  symbol: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
}

export interface WatchlistResponse {
  tickers: Ticker[];
}

// Live price update from WebSocket
export interface LivePrice {
  symbol: string;
  bid: number;
  ask: number;
  last: number;
  timestamp: string;
}

// WebSocket message types
export interface WSMessage {
  type: 'price' | 'status' | 'error';
  symbol?: string;
  bid?: number;
  ask?: number;
  last?: number;
  timestamp?: string;
  message?: string;
  connected?: boolean;
}

// Map of symbol to live price
export type LivePriceMap = Record<string, LivePrice>;
