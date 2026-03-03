// Market data types matching the backend models

export interface Gainer {
  ticker: string;
  price: number;
  changePercent: number;
  change: number;
  volume: number;
  previousClose: number;
  open: number;
  high: number;
  low: number;
  vwap: number;
  updatedAt: number;
}

export interface WebSocketMessage {
  type: 'gainers' | 'error' | 'connected';
  data: Gainer[];
  timestamp: string;
}

export interface ConnectionState {
  isConnected: boolean;
  error: string | null;
  reconnectAttempt: number;
}
